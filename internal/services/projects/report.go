package projects

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"idsai-core-up/internal/domain"
	"idsai-core-up/internal/services/projectflow"
	"idsai-core-up/internal/strutil"

	"github.com/google/uuid"
	"github.com/phpdave11/gofpdf"
)

var ErrFinalReportUnavailable = errors.New("final report available only for completed projects")

//go:embed fonts/Ubuntu-Regular.ttf
var ubuntuRegularTTF []byte

//go:embed fonts/Ubuntu-Bold.ttf
var ubuntuBoldTTF []byte

//go:embed fonts/JetBrainsMono-Regular.ttf
var jetBrainsMonoRegularTTF []byte

//go:embed fonts/JetBrainsMono-Bold.ttf
var jetBrainsMonoBoldTTF []byte

const (
	reportStorageSchemaVersion = 1
	reportMaxCriteria          = 50
	reportMaxTasks             = 100
	reportMaxTitleRunes        = 120
	reportMaxTextRunes         = 800
	reportMaxCommentRunes      = 400
	reportDefaultPlaceholder   = "-"
	reportDateTimeLayout       = "02.01.2006 15:04"
	reportDateLayout           = "02.01.2006"
	reportPageMargin           = 15.0
	reportPageWidth            = 180.0
	reportPageBottom           = 282.0
)

type reportColor struct {
	R int
	G int
	B int
}

var reportTheme = struct {
	BG         reportColor
	Panel      reportColor
	Surface    reportColor
	Line       reportColor
	Text       reportColor
	Muted      reportColor
	Dark       reportColor
	Blue       reportColor
	BlueSoft   reportColor
	Green      reportColor
	GreenSoft  reportColor
	Amber      reportColor
	AmberSoft  reportColor
	Danger     reportColor
	DangerSoft reportColor
}{
	BG:         reportColor{244, 246, 248},
	Panel:      reportColor{255, 255, 255},
	Surface:    reportColor{251, 252, 253},
	Line:       reportColor{228, 231, 236},
	Text:       reportColor{15, 23, 40},
	Muted:      reportColor{100, 116, 139},
	Dark:       reportColor{23, 26, 31},
	Blue:       reportColor{37, 99, 235},
	BlueSoft:   reportColor{239, 246, 255},
	Green:      reportColor{22, 163, 74},
	GreenSoft:  reportColor{236, 253, 243},
	Amber:      reportColor{245, 158, 11},
	AmberSoft:  reportColor{255, 247, 237},
	Danger:     reportColor{239, 68, 68},
	DangerSoft: reportColor{254, 242, 242},
}

type FinalReportSource interface {
	ListProjectStackCodes(ctx context.Context, projectID uuid.UUID) ([]string, error)
	ListProjectMembers(ctx context.Context, projectID uuid.UUID) ([]projectflow.Member, error)
	ListProjectCriteria(ctx context.Context, projectID uuid.UUID) ([]projectflow.Criterion, error)
	ListProjectCriterionGrades(ctx context.Context, projectID, professorID uuid.UUID) ([]projectflow.CriterionGrade, error)
	ListProjectTasks(ctx context.Context, projectID uuid.UUID) ([]projectflow.Task, error)
	ListProjectTaskActivities(ctx context.Context, projectID uuid.UUID, taskID *uuid.UUID) ([]projectflow.TaskActivity, error)
}

type FinalReportFile struct {
	Filename    string
	ContentType string
	Data        []byte
}

type finalReportArtifacts struct {
	Data         finalReportData
	PDFData      []byte
	SnapshotData []byte
}

type finalReportData struct {
	ProjectTitle       string
	ProjectDescription string
	ProjectID          string
	StatusLabel        string
	VisibilityLabel    string
	CreatedAtLabel     string
	UpdatedAtLabel     string
	GeneratedAtLabel   string
	LeaderLabel        string
	ReviewerLabel      string
	ReviewedAtLabel    string
	ScoreLabel         string
	PassPercentLabel   string
	CriteriaSummary    string
	RetakeSummary      string
	StacksLabel        string
	OverallComment     string
	Members            []finalReportMember
	Tasks              []finalReportTask
	Criteria           []finalReportCriterion
	TaskStats          finalReportTaskStats
	TasksTruncated     bool
	CriteriaTruncated  bool
}

type finalReportMember struct {
	Name   string
	Role   string
	Status string
}

type finalReportTask struct {
	Title       string
	Status      string
	Role        string
	Assignee    string
	Description string
}

type finalReportCriterion struct {
	Title       string
	Weight      string
	Result      string
	Description string
	Comment     string
}

type finalReportTaskStats struct {
	Total      int
	Open       int
	InProgress int
	Done       int
}

type finalReportKV struct {
	Label string
	Value string
}

func (s *Service) GetProjectFinalReportPDF(ctx context.Context, projectID, viewerID, viewerFacultyID uuid.UUID) (FinalReportFile, error) {
	view, err := s.getAccessibleFinalReportView(ctx, projectID, viewerID, viewerFacultyID)
	if err != nil {
		return FinalReportFile{}, err
	}

	if raw, ok := s.loadStoredFinalReportPDF(ctx, view.Project); ok {
		return FinalReportFile{
			Filename:    reportFilename(view.Project.ID),
			ContentType: "application/pdf",
			Data:        raw,
		}, nil
	}

	artifacts, err := s.buildFinalReportArtifacts(ctx, view)
	if err != nil {
		return FinalReportFile{}, err
	}
	if err := s.storeFinalReportArtifacts(ctx, view.Project, artifacts); err != nil && !errors.Is(err, ErrStorage) {
		log.Printf("projects final report store skipped project_id=%s err=%v", view.Project.ID, err)
	}

	return FinalReportFile{
		Filename:    reportFilename(view.Project.ID),
		ContentType: "application/pdf",
		Data:        artifacts.PDFData,
	}, nil
}

func (s *Service) CaptureProjectFinalReport(ctx context.Context, projectID, viewerID, viewerFacultyID uuid.UUID) error {
	view, err := s.getAccessibleFinalReportView(ctx, projectID, viewerID, viewerFacultyID)
	if err != nil {
		return err
	}
	artifacts, err := s.buildFinalReportArtifacts(ctx, view)
	if err != nil {
		return err
	}
	return s.storeFinalReportArtifacts(ctx, view.Project, artifacts)
}

func (s *Service) getAccessibleFinalReportView(ctx context.Context, projectID, viewerID, viewerFacultyID uuid.UUID) (ProjectView, error) {
	view, err := s.GetProjectViewForViewer(ctx, projectID, viewerID, viewerFacultyID)
	if err != nil {
		return ProjectView{}, err
	}
	if view.Project.Status != domain.ProjectCompleted && view.Project.Status != domain.ProjectArchive {
		return ProjectView{}, ErrFinalReportUnavailable
	}
	if !view.Access.CanViewFinalGrade {
		return ProjectView{}, domain.ErrForbidden
	}
	return view, nil
}

func (s *Service) buildFinalReportArtifacts(ctx context.Context, view ProjectView) (finalReportArtifacts, error) {
	data, err := s.buildFinalReportData(ctx, view)
	if err != nil {
		return finalReportArtifacts{}, err
	}
	pdfRaw, err := renderFinalReportPDF(data)
	if err != nil {
		return finalReportArtifacts{}, err
	}
	snapshotRaw, err := json.MarshalIndent(finalReportSnapshotPayload(data), "", "  ")
	if err != nil {
		return finalReportArtifacts{}, err
	}
	return finalReportArtifacts{
		Data:         data,
		PDFData:      pdfRaw,
		SnapshotData: snapshotRaw,
	}, nil
}

func (s *Service) buildFinalReportData(ctx context.Context, view ProjectView) (finalReportData, error) {
	if s.reportSource == nil {
		return finalReportData{}, ErrReportSource
	}

	project := view.Project
	projectID := project.ID

	stacks, err := s.reportSource.ListProjectStackCodes(ctx, projectID)
	if err != nil {
		return finalReportData{}, err
	}
	members, err := s.reportSource.ListProjectMembers(ctx, projectID)
	if err != nil {
		return finalReportData{}, err
	}
	criteria, err := s.reportSource.ListProjectCriteria(ctx, projectID)
	if err != nil {
		return finalReportData{}, err
	}

	var grades []projectflow.CriterionGrade
	if project.ProfessorID != nil {
		grades, err = s.reportSource.ListProjectCriterionGrades(ctx, projectID, *project.ProfessorID)
		if err != nil {
			return finalReportData{}, err
		}
	}

	tasks, err := s.reportSource.ListProjectTasks(ctx, projectID)
	if err != nil {
		return finalReportData{}, err
	}

	return newFinalReportData(project, view.ReviewSummary, stacks, members, criteria, grades, tasks), nil
}

func (s *Service) storeFinalReportArtifacts(ctx context.Context, project domain.Project, artifacts finalReportArtifacts) error {
	if s.storage == nil || !s.storage.Available() {
		return ErrStorage
	}

	jsonKey := finalReportJSONKey(project.ID, project.RetakeCount)
	if err := s.storage.PutObject(ctx, jsonKey, "application/json; charset=utf-8", artifacts.SnapshotData); err != nil {
		return ErrStorage
	}
	pdfKey := finalReportPDFKey(project.ID, project.RetakeCount)
	if err := s.storage.PutObject(ctx, pdfKey, "application/pdf", artifacts.PDFData); err != nil {
		_ = s.storage.DeleteObject(ctx, jsonKey)
		return ErrStorage
	}
	return nil
}

func (s *Service) loadStoredFinalReportPDF(ctx context.Context, project domain.Project) ([]byte, bool) {
	if s.storage == nil || !s.storage.Available() {
		return nil, false
	}
	raw, err := s.storage.GetObject(ctx, finalReportPDFKey(project.ID, project.RetakeCount))
	if err != nil || len(raw) < 4 || !bytes.HasPrefix(raw, []byte("%PDF")) {
		return nil, false
	}
	return raw, true
}

func finalReportPDFKey(projectID uuid.UUID, retakeCount int) string {
	return fmt.Sprintf("projects/final-reports/%s/retake-%02d.pdf", projectID.String(), max(0, retakeCount))
}

func finalReportJSONKey(projectID uuid.UUID, retakeCount int) string {
	return fmt.Sprintf("projects/final-reports/%s/retake-%02d.json", projectID.String(), max(0, retakeCount))
}

func reportFilename(projectID uuid.UUID) string {
	return fmt.Sprintf("project-report-%s.pdf", projectID.String())
}

func finalReportSnapshotPayload(data finalReportData) map[string]any {
	return map[string]any{
		"version":      reportStorageSchemaVersion,
		"generated_at": data.GeneratedAtLabel,
		"project": map[string]any{
			"id":          data.ProjectID,
			"title":       data.ProjectTitle,
			"description": data.ProjectDescription,
			"status":      data.StatusLabel,
			"visibility":  data.VisibilityLabel,
			"created_at":  data.CreatedAtLabel,
			"updated_at":  data.UpdatedAtLabel,
			"stacks":      data.StacksLabel,
		},
		"summary": map[string]any{
			"score":        data.ScoreLabel,
			"pass_percent": data.PassPercentLabel,
			"criteria":     data.CriteriaSummary,
			"retake":       data.RetakeSummary,
			"leader":       data.LeaderLabel,
			"reviewer":     data.ReviewerLabel,
			"reviewed_at":  data.ReviewedAtLabel,
			"overall":      data.OverallComment,
			"task_stats":   data.TaskStats,
			"tasks_cut":    data.TasksTruncated,
			"criteria_cut": data.CriteriaTruncated,
		},
		"members":  data.Members,
		"tasks":    data.Tasks,
		"criteria": data.Criteria,
	}
}

func newFinalReportData(
	project domain.Project,
	summary *ReviewSummary,
	stacks []string,
	members []projectflow.Member,
	criteria []projectflow.Criterion,
	grades []projectflow.CriterionGrade,
	tasks []projectflow.Task,
) finalReportData {
	leader := choosePersonLabel(project.CreatedByName, project.CreatedByEmail, project.CreatedBy.String())
	reviewer := reportDefaultPlaceholder
	reviewedAt := reportDefaultPlaceholder
	score := "не опубликована"
	passPercent := reportDefaultPlaceholder
	criteriaSummary := "0 / 0"
	retakeSummary := fmt.Sprintf("%d пересдач, штраф %d%%", max(0, project.RetakeCount), domain.RetakePenaltyPercent(project.RetakeCount))
	if summary != nil {
		reviewer = fallbackString(strings.TrimSpace(summary.Reviewer), reportDefaultPlaceholder)
		if summary.ReviewedAt != nil {
			reviewedAt = formatDateTime(*summary.ReviewedAt)
		}
		score = fallbackString(strings.TrimSpace(summary.Score), score)
		passPercent = fmt.Sprintf("%d%%", summary.PassPercent)
		criteriaSummary = fmt.Sprintf("%d / %d", summary.Met, summary.Total)
	}

	memberNameByUserID := map[string]string{
		project.CreatedBy.String(): leader,
	}
	reportMembers := make([]finalReportMember, 0, len(members)+1)
	reportMembers = append(reportMembers, finalReportMember{
		Name:   leader,
		Role:   "Лидер проекта",
		Status: "ACTIVE",
	})
	for _, item := range members {
		name := choosePersonLabel(item.FullName, item.Email, item.UserID)
		memberNameByUserID[item.UserID] = name
		role := fallbackString(strings.TrimSpace(firstNonEmptyPtr(item.PositionName, item.PositionCode)), "Участник")
		reportMembers = append(reportMembers, finalReportMember{
			Name:   name,
			Role:   role,
			Status: strings.ToUpper(strings.TrimSpace(item.Status)),
		})
	}
	sort.SliceStable(reportMembers[1:], func(i, j int) bool {
		left := reportMembers[i+1]
		right := reportMembers[j+1]
		if left.Status != right.Status {
			if left.Status == "ACTIVE" {
				return true
			}
			if right.Status == "ACTIVE" {
				return false
			}
			return left.Status < right.Status
		}
		return left.Name < right.Name
	})

	gradeByCriterion := make(map[string]projectflow.CriterionGrade, len(grades))
	for _, item := range grades {
		gradeByCriterion[strings.TrimSpace(item.CriterionID)] = item
	}

	criteriaTruncated := len(criteria) > reportMaxCriteria
	if criteriaTruncated {
		criteria = criteria[:reportMaxCriteria]
	}
	reportCriteria := make([]finalReportCriterion, 0, len(criteria))
	commentCandidates := make([]finalReportCriterion, 0, len(criteria))
	for _, item := range criteria {
		grade, ok := gradeByCriterion[strings.TrimSpace(item.ID)]
		result := "Не проверено"
		if ok && grade.IsMet != nil {
			if *grade.IsMet {
				result = "Выполнено"
			} else {
				result = "Есть замечания"
			}
		}
		row := finalReportCriterion{
			Title:       safeReportText(item.Title, reportMaxTitleRunes),
			Weight:      fmt.Sprintf("%d", max(1, item.Weight)),
			Result:      result,
			Description: safeReportText(item.Description, reportMaxTextRunes),
			Comment:     safeReportText(grade.Comment, reportMaxCommentRunes),
		}
		reportCriteria = append(reportCriteria, row)
		if row.Comment != reportDefaultPlaceholder {
			commentCandidates = append(commentCandidates, row)
		}
	}
	sort.SliceStable(commentCandidates, func(i, j int) bool {
		left := 1
		if commentCandidates[i].Result == "Есть замечания" {
			left = 0
		}
		right := 1
		if commentCandidates[j].Result == "Есть замечания" {
			right = 0
		}
		if left != right {
			return left < right
		}
		return commentCandidates[i].Title < commentCandidates[j].Title
	})

	overallComment := "Комментарий преподавателя пока не добавлен."
	if len(commentCandidates) > 0 {
		parts := make([]string, 0, min(3, len(commentCandidates)))
		for _, item := range commentCandidates[:min(3, len(commentCandidates))] {
			parts = append(parts, fmt.Sprintf("%s: %s", item.Title, item.Comment))
		}
		overallComment = strings.Join(parts, "\n\n")
	} else if summary != nil && summary.Total > 0 && summary.Met == summary.Total {
		overallComment = "Все критерии отмечены как выполненные."
	} else if summary != nil && summary.Total > 0 {
		overallComment = fmt.Sprintf("Есть замечания по %d критериям.", max(0, summary.Total-summary.Met))
	}

	reportTasks := make([]finalReportTask, 0, min(len(tasks), reportMaxTasks))
	taskStats := finalReportTaskStats{Total: len(tasks)}
	for idx, item := range tasks {
		status := strings.ToUpper(strings.TrimSpace(item.Status))
		switch status {
		case "DONE":
			taskStats.Done++
		case "IN_PROGRESS":
			taskStats.InProgress++
		default:
			taskStats.Open++
		}
		if idx >= reportMaxTasks {
			continue
		}
		assignee := reportDefaultPlaceholder
		if item.AssigneeUserID != nil && strings.TrimSpace(*item.AssigneeUserID) != "" {
			assignee = fallbackString(strings.TrimSpace(memberNameByUserID[strings.TrimSpace(*item.AssigneeUserID)]), strings.TrimSpace(*item.AssigneeUserID))
		}
		reportTasks = append(reportTasks, finalReportTask{
			Title:       safeReportText(item.Title, reportMaxTitleRunes),
			Status:      taskStatusLabel(status),
			Role:        fallbackString(strings.TrimSpace(item.PositionName), strings.TrimSpace(item.PositionCode)),
			Assignee:    fallbackString(assignee, reportDefaultPlaceholder),
			Description: safeReportText(item.Description, reportMaxCommentRunes),
		})
	}

	return finalReportData{
		ProjectTitle:       safeReportText(project.Title, reportMaxTitleRunes),
		ProjectDescription: safeReportText(project.Description, reportMaxTextRunes),
		ProjectID:          project.ID.String(),
		StatusLabel:        projectStatusLabel(project.Status),
		VisibilityLabel:    visibilityLabel(project.IsPublic, project.Visibility),
		CreatedAtLabel:     formatDateTime(project.CreatedAt),
		UpdatedAtLabel:     formatDateTime(project.UpdatedAt),
		GeneratedAtLabel:   formatDateTime(time.Now().UTC()),
		LeaderLabel:        leader,
		ReviewerLabel:      reviewer,
		ReviewedAtLabel:    reviewedAt,
		ScoreLabel:         score,
		PassPercentLabel:   passPercent,
		CriteriaSummary:    criteriaSummary,
		RetakeSummary:      retakeSummary,
		StacksLabel:        formatStacks(stacks),
		OverallComment:     safeReportText(overallComment, reportMaxTextRunes),
		Members:            reportMembers,
		Tasks:              reportTasks,
		Criteria:           reportCriteria,
		TaskStats:          taskStats,
		TasksTruncated:     len(tasks) > reportMaxTasks,
		CriteriaTruncated:  criteriaTruncated,
	}
}

func renderFinalReportPDF(data finalReportData) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetCompression(false)
	pdf.SetMargins(reportPageMargin, 18, reportPageMargin)
	pdf.SetAutoPageBreak(true, 15)
	pdf.AddUTF8FontFromBytes("sans", "", ubuntuRegularTTF)
	pdf.AddUTF8FontFromBytes("sans", "B", ubuntuBoldTTF)
	pdf.AddUTF8FontFromBytes("mono", "", jetBrainsMonoRegularTTF)
	pdf.AddUTF8FontFromBytes("mono", "B", jetBrainsMonoBoldTTF)
	pdf.SetTitle(data.ProjectTitle, false)
	pdf.SetAuthor("IDSAI Core", false)
	pdf.SetSubject("Final project report", false)
	pdf.SetHeaderFuncMode(func() {
		reportPageChrome(pdf)
	}, true)
	pdf.SetFooterFunc(func() {
		reportFooter(pdf, data.ProjectID)
	})
	pdf.AddPage()

	reportHero(pdf, data)
	reportStatsGrid(pdf, data)

	reportSectionHeading(pdf, "PROJECT", "Паспорт проекта")
	reportKeyValueCard(pdf, "", []finalReportKV{
		{Label: "Лидер проекта", Value: data.LeaderLabel},
		{Label: "Проверяющий", Value: data.ReviewerLabel},
		{Label: "Дата проверки", Value: data.ReviewedAtLabel},
		{Label: "Создан", Value: data.CreatedAtLabel},
		{Label: "Обновлен", Value: data.UpdatedAtLabel},
		{Label: "Технологический стек", Value: data.StacksLabel},
	})

	reportSectionHeading(pdf, "OVERVIEW", "Описание")
	reportTextCard(pdf, "", data.ProjectDescription)

	reportSectionHeading(pdf, "REVIEW", "Комментарий преподавателя")
	reportTextCard(pdf, "", data.OverallComment)

	reportSectionHeading(pdf, "TEAM", "Команда")
	for _, item := range data.Members {
		reportItemCard(pdf, item.Name, item.Status, reportStatusTone(item.Status), item.Role, "")
	}

	reportSectionHeading(pdf, "TASKS", "Задачи и вклад")
	reportTextNote(pdf, fmt.Sprintf("Всего задач: %d. Выполнено: %d. В работе: %d. Открыто: %d.", data.TaskStats.Total, data.TaskStats.Done, data.TaskStats.InProgress, data.TaskStats.Open))
	if data.TasksTruncated {
		reportTextNote(pdf, fmt.Sprintf("Показаны первые %d задач, чтобы отчёт оставался компактным.", reportMaxTasks))
	}
	if len(data.Tasks) == 0 {
		reportTextCard(pdf, "", "Задачи в проекте не найдены.")
	} else {
		for _, item := range data.Tasks {
			reportItemCard(
				pdf,
				item.Title,
				item.Status,
				reportTaskTone(item.Status),
				fmt.Sprintf("Роль: %s | Исполнитель: %s", fallbackString(item.Role, reportDefaultPlaceholder), item.Assignee),
				item.Description,
			)
		}
	}

	reportSectionHeading(pdf, "CRITERIA", "Критерии и результат")
	if data.CriteriaTruncated {
		reportTextNote(pdf, fmt.Sprintf("Показаны первые %d критериев.", reportMaxCriteria))
	}
	if len(data.Criteria) == 0 {
		reportTextCard(pdf, "", "Критерии для этого проекта не найдены.")
	} else {
		for _, item := range data.Criteria {
			body := item.Description
			if item.Comment != reportDefaultPlaceholder {
				body = fallbackBody(body, "Комментарий: "+item.Comment)
			}
			reportItemCard(
				pdf,
				item.Title,
				item.Result,
				reportCriterionTone(item.Result),
				fmt.Sprintf("Вес: %s", item.Weight),
				body,
			)
		}
	}

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func reportPageChrome(pdf *gofpdf.Fpdf) {
	setFill(pdf, reportTheme.BG)
	pdf.Rect(0, 0, 210, 297, "F")
	setFill(pdf, reportTheme.Dark)
	pdf.Rect(0, 0, 210, 5, "F")
	setDraw(pdf, reportTheme.Line)
	pdf.Line(reportPageMargin, 14, 210-reportPageMargin, 14)
}

func reportFooter(pdf *gofpdf.Fpdf, projectID string) {
	pdf.SetY(-12)
	pdf.SetFont("mono", "", 7.5)
	setText(pdf, reportTheme.Muted)
	pdf.CellFormat(0, 5, fmt.Sprintf("project %s | page %d", strutil.TruncateUTF8(projectID, 20), pdf.PageNo()), "", 0, "C", false, 0, "")
}

func reportHero(pdf *gofpdf.Fpdf, data finalReportData) {
	x := reportPageMargin
	y := pdf.GetY()
	w := reportPageWidth
	h := 46.0

	ensureReportSpace(pdf, h+8)
	setFill(pdf, reportTheme.Dark)
	setDraw(pdf, reportTheme.Dark)
	pdf.RoundedRect(x, y, w, h, 4, "1234", "FD")

	scoreW := 42.0
	scoreH := 26.0
	scoreX := x + w - scoreW - 6
	scoreY := y + 10
	setFill(pdf, reportTheme.Blue)
	setDraw(pdf, reportTheme.Blue)
	pdf.RoundedRect(scoreX, scoreY, scoreW, scoreH, 3, "1234", "FD")

	pdf.SetXY(x+7, y+7)
	pdf.SetFont("mono", "B", 8)
	setText(pdf, reportTheme.BlueSoft)
	pdf.CellFormat(0, 4, "IDSAI FINAL REPORT", "", 1, "L", false, 0, "")

	pdf.SetX(x + 7)
	pdf.SetFont("sans", "B", 19)
	setTextRGB(pdf, 255, 255, 255)
	pdf.MultiCell(w-scoreW-24, 8.5, data.ProjectTitle, "", "L", false)

	pdf.SetX(x + 7)
	pdf.SetFont("sans", "", 9.5)
	setTextRGB(pdf, 214, 221, 232)
	pdf.MultiCell(w-scoreW-24, 4.8, "Финальный срез проекта после публикации преподавательской оценки.", "", "L", false)

	pdf.SetXY(x+7, y+h-11)
	pdf.SetFont("mono", "", 7.8)
	setTextRGB(pdf, 214, 221, 232)
	pdf.CellFormat(0, 4, fmt.Sprintf("%s | %s | %s", data.StatusLabel, data.VisibilityLabel, data.GeneratedAtLabel), "", 1, "L", false, 0, "")

	pdf.SetXY(scoreX, scoreY+4)
	pdf.SetFont("mono", "", 7.2)
	setTextRGB(pdf, 214, 234, 255)
	pdf.CellFormat(scoreW, 4, "ИТОГ", "", 1, "C", false, 0, "")
	pdf.SetX(scoreX)
	pdf.SetFont("sans", "B", 16)
	setTextRGB(pdf, 255, 255, 255)
	pdf.CellFormat(scoreW, 8, data.ScoreLabel, "", 1, "C", false, 0, "")
	pdf.SetX(scoreX)
	pdf.SetFont("mono", "", 7.2)
	setTextRGB(pdf, 214, 234, 255)
	pdf.CellFormat(scoreW, 4, data.PassPercentLabel, "", 1, "C", false, 0, "")

	pdf.SetY(y + h + 7)
}

func reportStatsGrid(pdf *gofpdf.Fpdf, data finalReportData) {
	const gap = 6.0
	cardW := (reportPageWidth - gap) / 2
	cardH := 23.0
	x := reportPageMargin
	y := pdf.GetY()
	ensureReportSpace(pdf, cardH*2+gap+6)

	reportStatCard(pdf, x, y, cardW, cardH, "Критерии", data.CriteriaSummary, "green")
	reportStatCard(pdf, x+cardW+gap, y, cardW, cardH, "Пересдачи", data.RetakeSummary, "amber")
	reportStatCard(pdf, x, y+cardH+gap, cardW, cardH, "Завершено задач", fmt.Sprintf("%d / %d", data.TaskStats.Done, data.TaskStats.Total), "blue")
	reportStatCard(pdf, x+cardW+gap, y+cardH+gap, cardW, cardH, "Проверяющий", data.ReviewerLabel, "muted")

	pdf.SetY(y + cardH*2 + gap + 6)
}

func reportStatCard(pdf *gofpdf.Fpdf, x, y, w, h float64, label, value, tone string) {
	fill, border, valueColor := reportToneColors(tone)
	setFill(pdf, reportTheme.Panel)
	setDraw(pdf, reportTheme.Line)
	pdf.RoundedRect(x, y, w, h, 3, "1234", "FD")

	setFill(pdf, fill)
	pdf.RoundedRect(x+4, y+4, 18, h-8, 2.4, "1234", "F")

	pdf.SetXY(x+26, y+4)
	pdf.SetFont("mono", "", 7.5)
	setText(pdf, reportTheme.Muted)
	pdf.CellFormat(w-30, 4, strings.ToUpper(label), "", 1, "L", false, 0, "")

	pdf.SetX(x + 26)
	pdf.SetFont("sans", "B", 11)
	setText(pdf, valueColor)
	pdf.MultiCell(w-30, 5.2, value, "", "L", false)
	_ = border
}

func reportSectionHeading(pdf *gofpdf.Fpdf, label, title string) {
	ensureReportSpace(pdf, 12)
	pdf.SetFont("mono", "", 7.5)
	setText(pdf, reportTheme.Blue)
	pdf.CellFormat(0, 4, label, "", 1, "L", false, 0, "")
	pdf.SetFont("sans", "B", 15)
	setText(pdf, reportTheme.Text)
	pdf.CellFormat(0, 6, title, "", 1, "L", false, 0, "")
	setDraw(pdf, reportTheme.Line)
	pdf.Line(reportPageMargin, pdf.GetY()+1, reportPageMargin+26, pdf.GetY()+1)
	pdf.Ln(4)
}

func reportKeyValueCard(pdf *gofpdf.Fpdf, title string, rows []finalReportKV) {
	contentW := reportPageWidth - 12
	height := 12.0
	if title != "" {
		height += 6
	}
	for _, row := range rows {
		valueLines := reportSplitText(pdf, "sans", "", 10, row.Value, contentW)
		height += 4 + float64(max(1, len(valueLines)))*5 + 2
	}
	reportContainerCard(pdf, height, func(x, y, w float64) {
		cursorY := y + 6
		if title != "" {
			pdf.SetXY(x+6, cursorY)
			pdf.SetFont("sans", "B", 12)
			setText(pdf, reportTheme.Text)
			pdf.CellFormat(w-12, 5, title, "", 1, "L", false, 0, "")
			cursorY += 7
		}
		for _, row := range rows {
			pdf.SetXY(x+6, cursorY)
			pdf.SetFont("mono", "", 7.5)
			setText(pdf, reportTheme.Muted)
			pdf.CellFormat(w-12, 4, strings.ToUpper(row.Label), "", 1, "L", false, 0, "")

			pdf.SetXY(x+6, cursorY+4)
			pdf.SetFont("sans", "", 10)
			setText(pdf, reportTheme.Text)
			pdf.MultiCell(w-12, 5, row.Value, "", "L", false)
			cursorY = pdf.GetY() + 2
		}
	})
}

func reportTextCard(pdf *gofpdf.Fpdf, title, body string) {
	bodyLines := reportSplitText(pdf, "sans", "", 10, body, reportPageWidth-12)
	height := 12.0 + float64(max(1, len(bodyLines)))*5
	if title != "" {
		height += 7
	}
	reportContainerCard(pdf, height, func(x, y, w float64) {
		cursorY := y + 6
		if title != "" {
			pdf.SetXY(x+6, cursorY)
			pdf.SetFont("sans", "B", 12)
			setText(pdf, reportTheme.Text)
			pdf.CellFormat(w-12, 5, title, "", 1, "L", false, 0, "")
			cursorY += 7
		}
		pdf.SetXY(x+6, cursorY)
		pdf.SetFont("sans", "", 10)
		setText(pdf, reportTheme.Text)
		pdf.MultiCell(w-12, 5, body, "", "L", false)
	})
}

func reportTextNote(pdf *gofpdf.Fpdf, text string) {
	pdf.SetFont("sans", "", 9)
	setText(pdf, reportTheme.Muted)
	pdf.MultiCell(reportPageWidth, 4.5, text, "", "L", false)
	pdf.Ln(1.5)
}

func reportItemCard(pdf *gofpdf.Fpdf, title, badge, tone, meta, body string) {
	const padding = 6.0

	badgeWidth := 0.0
	if badge != "" {
		pdf.SetFont("mono", "B", 7.2)
		badgeWidth = min(58, pdf.GetStringWidth(strings.ToUpper(badge))+8)
	}

	titleWidth := reportPageWidth - padding*2
	if badgeWidth > 0 {
		titleWidth -= badgeWidth + 4
	}
	titleLines := reportSplitText(pdf, "sans", "B", 11, title, titleWidth)
	metaLines := reportSplitText(pdf, "mono", "", 8.2, meta, reportPageWidth-padding*2)
	bodyLines := []string{}
	if strings.TrimSpace(body) != "" {
		bodyLines = reportSplitText(pdf, "sans", "", 9.5, body, reportPageWidth-padding*2)
	}

	height := padding + float64(max(1, len(titleLines)))*5
	if len(metaLines) > 0 {
		height += 2 + float64(len(metaLines))*4.1
	}
	if len(bodyLines) > 0 {
		height += 2 + float64(len(bodyLines))*4.8
	}
	height += 6

	reportContainerCard(pdf, height, func(x, y, w float64) {
		cursorY := y + padding

		if badgeWidth > 0 {
			reportBadge(pdf, x+w-padding-badgeWidth, cursorY-0.5, badgeWidth, badge, tone)
		}

		pdf.SetXY(x+padding, cursorY)
		pdf.SetFont("sans", "B", 11)
		setText(pdf, reportTheme.Text)
		pdf.MultiCell(titleWidth, 5, title, "", "L", false)
		cursorY = pdf.GetY()

		if len(metaLines) > 0 {
			pdf.SetXY(x+padding, cursorY+1)
			pdf.SetFont("mono", "", 8.2)
			setText(pdf, reportTheme.Muted)
			pdf.MultiCell(w-padding*2, 4.1, meta, "", "L", false)
			cursorY = pdf.GetY()
		}

		if len(bodyLines) > 0 {
			pdf.SetXY(x+padding, cursorY+2)
			pdf.SetFont("sans", "", 9.5)
			setText(pdf, reportTheme.Text)
			pdf.MultiCell(w-padding*2, 4.8, body, "", "L", false)
		}
	})
}

func reportContainerCard(pdf *gofpdf.Fpdf, height float64, drawContent func(x, y, w float64)) {
	ensureReportSpace(pdf, height+4)
	x := reportPageMargin
	y := pdf.GetY()
	setFill(pdf, reportTheme.Panel)
	setDraw(pdf, reportTheme.Line)
	pdf.RoundedRect(x, y, reportPageWidth, height, 3.5, "1234", "FD")
	drawContent(x, y, reportPageWidth)
	pdf.SetY(y + height + 4)
}

func reportBadge(pdf *gofpdf.Fpdf, x, y, w float64, label, tone string) {
	fill, _, text := reportToneColors(tone)
	setFill(pdf, fill)
	setDraw(pdf, fill)
	pdf.RoundedRect(x, y, w, 6, 2, "1234", "FD")
	pdf.SetXY(x, y+1.1)
	pdf.SetFont("mono", "B", 7.2)
	setText(pdf, text)
	pdf.CellFormat(w, 3.8, strings.ToUpper(label), "", 1, "C", false, 0, "")
}

func reportToneColors(tone string) (fill, border, text reportColor) {
	switch strings.ToLower(strings.TrimSpace(tone)) {
	case "green":
		return reportTheme.GreenSoft, reportTheme.GreenSoft, reportTheme.Green
	case "amber":
		return reportTheme.AmberSoft, reportTheme.AmberSoft, reportTheme.Amber
	case "danger":
		return reportTheme.DangerSoft, reportTheme.DangerSoft, reportTheme.Danger
	case "blue":
		return reportTheme.BlueSoft, reportTheme.BlueSoft, reportTheme.Blue
	default:
		return reportTheme.Surface, reportTheme.Surface, reportTheme.Muted
	}
}

func reportStatusTone(status string) string {
	switch strings.ToUpper(strings.TrimSpace(status)) {
	case "ACTIVE":
		return "green"
	case "INVITED", "APPLIED":
		return "amber"
	case "REJECTED", "REMOVED":
		return "danger"
	default:
		return "muted"
	}
}

func reportTaskTone(status string) string {
	switch strings.ToUpper(strings.TrimSpace(status)) {
	case "ЗАВЕРШЕНА":
		return "green"
	case "В РАБОТЕ":
		return "blue"
	default:
		return "amber"
	}
}

func reportCriterionTone(result string) string {
	switch strings.ToUpper(strings.TrimSpace(result)) {
	case "ВЫПОЛНЕНО":
		return "green"
	case "ЕСТЬ ЗАМЕЧАНИЯ":
		return "danger"
	default:
		return "amber"
	}
}

func reportSplitText(pdf *gofpdf.Fpdf, family, style string, size float64, text string, width float64) []string {
	pdf.SetFont(family, style, size)
	lines := pdf.SplitText(fallbackString(text, reportDefaultPlaceholder), width)
	if len(lines) == 0 {
		return []string{reportDefaultPlaceholder}
	}
	return lines
}

func ensureReportSpace(pdf *gofpdf.Fpdf, needed float64) {
	if pdf.GetY()+needed <= reportPageBottom {
		return
	}
	pdf.AddPage()
}

func setFill(pdf *gofpdf.Fpdf, c reportColor) {
	pdf.SetFillColor(c.R, c.G, c.B)
}

func setDraw(pdf *gofpdf.Fpdf, c reportColor) {
	pdf.SetDrawColor(c.R, c.G, c.B)
}

func setText(pdf *gofpdf.Fpdf, c reportColor) {
	pdf.SetTextColor(c.R, c.G, c.B)
}

func setTextRGB(pdf *gofpdf.Fpdf, r, g, b int) {
	pdf.SetTextColor(r, g, b)
}

func fallbackBody(parts ...string) string {
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" || part == reportDefaultPlaceholder {
			continue
		}
		out = append(out, part)
	}
	if len(out) == 0 {
		return ""
	}
	return strings.Join(out, "\n\n")
}

func safeReportText(value string, maxRunes int) string {
	value = strings.ReplaceAll(value, "\r\n", "\n")
	value = strings.ReplaceAll(value, "\r", "\n")
	value = strings.TrimSpace(value)
	if value == "" {
		return reportDefaultPlaceholder
	}
	return strutil.TruncateUTF8(value, maxRunes)
}

func choosePersonLabel(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return strutil.TruncateUTF8(value, reportMaxTitleRunes)
		}
	}
	return reportDefaultPlaceholder
}

func fallbackString(value, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	return value
}

func formatStacks(items []string) string {
	clean := make([]string, 0, len(items))
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		clean = append(clean, strutil.TruncateUTF8(item, 40))
	}
	if len(clean) == 0 {
		return "Не указан"
	}
	return strings.Join(clean, ", ")
}

func projectStatusLabel(status domain.ProjectStatus) string {
	switch status {
	case domain.ProjectDraft:
		return "Черновик"
	case domain.ProjectReview:
		return "Подготовка"
	case domain.ProjectRecruitment:
		return "Набор команды"
	case domain.ProjectActive:
		return "В работе"
	case domain.ProjectGrading:
		return "На оценивании"
	case domain.ProjectCompleted:
		return "Завершен"
	case domain.ProjectArchive:
		return "Архив"
	default:
		return string(status)
	}
}

func visibilityLabel(isPublic bool, visibility string) string {
	if isPublic {
		return "Публичный"
	}
	visibility = strings.ToUpper(strings.TrimSpace(visibility))
	if visibility == "" {
		return "Приватный"
	}
	return visibility
}

func taskStatusLabel(status string) string {
	switch strings.ToUpper(strings.TrimSpace(status)) {
	case "DONE":
		return "Завершена"
	case "IN_PROGRESS":
		return "В работе"
	default:
		return "Открыта"
	}
}

func formatDateTime(value time.Time) string {
	if value.IsZero() {
		return reportDefaultPlaceholder
	}
	if value.Hour() == 0 && value.Minute() == 0 && value.Second() == 0 {
		return value.Format(reportDateLayout)
	}
	return value.Format(reportDateTimeLayout)
}

func firstNonEmptyPtr(values ...*string) string {
	for _, value := range values {
		if value == nil {
			continue
		}
		trimmed := strings.TrimSpace(*value)
		if trimmed != "" {
			return trimmed
		}
	}
	return ""
}
