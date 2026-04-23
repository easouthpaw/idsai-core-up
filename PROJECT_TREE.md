# Project Tree

Generated: 2026-03-31 22:52:08
Excluded: `.git`
Total: 51 directories, 250 files

```text
idsai-core-up
|-- .githooks
|   `-- pre-push
|-- .github
|   `-- workflows
|       `-- ci.yml
|-- .vscode
|   `-- settings.json
|-- cmd
|   `-- api
|       `-- main.go
|-- docs
|   |-- architecture
|   |   `-- ARCHITECTURE.md
|   |-- swagger
|   |   |-- docs.go
|   |   |-- swagger.json
|   |   `-- swagger.yaml
|   `-- PROJECT_LIFECYCLE.md
|-- internal
|   |-- app
|   |   |-- app.go
|   |   |-- wire_modules.go
|   |   `-- wire_outbox.go
|   |-- architecture
|   |   `-- dependency_rules_test.go
|   |-- config
|   |   `-- config.go
|   |-- db
|   |   |-- db.go
|   |   `-- db_integration_test.go
|   |-- domain
|   |   |-- errors.go
|   |   |-- kb.go
|   |   |-- member.go
|   |   |-- position.go
|   |   |-- project.go
|   |   `-- task.go
|   |-- http
|   |   |-- frontend
|   |   |   |-- assets
|   |   |   |   |-- author-rabbit.png
|   |   |   |   |-- code-screen.jpg
|   |   |   |   |-- enu-logo.jpg
|   |   |   |   |-- idsai-corp-logo.png
|   |   |   |   |-- IDSAI-Corp.png
|   |   |   |   `-- terminal-2.svg
|   |   |   |-- css
|   |   |   |   |-- admin.css
|   |   |   |   |-- auth.css
|   |   |   |   |-- author.css
|   |   |   |   |-- dev-theme.css
|   |   |   |   |-- groups.css
|   |   |   |   |-- invites.css
|   |   |   |   |-- kb.css
|   |   |   |   |-- landing.css
|   |   |   |   |-- not-found.css
|   |   |   |   |-- notifications-popup.css
|   |   |   |   |-- professor.css
|   |   |   |   |-- profile.css
|   |   |   |   |-- project-detail.css
|   |   |   |   |-- projects.css
|   |   |   |   |-- role-sidebar.css
|   |   |   |   `-- settings.css
|   |   |   |-- js
|   |   |   |   |-- admin.js
|   |   |   |   |-- auth-session.js
|   |   |   |   |-- author.js
|   |   |   |   |-- groups.js
|   |   |   |   |-- invites.js
|   |   |   |   |-- kb-article.js
|   |   |   |   |-- kb.js
|   |   |   |   |-- landing.js
|   |   |   |   |-- login.js
|   |   |   |   |-- not-found.js
|   |   |   |   |-- notifications-popup.js
|   |   |   |   |-- professor-criteria.js
|   |   |   |   |-- professor-grading.js
|   |   |   |   |-- professor.js
|   |   |   |   |-- profile.js
|   |   |   |   |-- project-detail.js
|   |   |   |   |-- projects.js
|   |   |   |   |-- role-sidebar.js
|   |   |   |   `-- settings.js
|   |   |   |-- 404.html
|   |   |   |-- admin.html
|   |   |   |-- author.html
|   |   |   |-- frontend.go
|   |   |   |-- groups.html
|   |   |   |-- invites.html
|   |   |   |-- kb-article.html
|   |   |   |-- kb.html
|   |   |   |-- landing.html
|   |   |   |-- login.html
|   |   |   |-- professor-criteria.html
|   |   |   |-- professor-grading.html
|   |   |   |-- professor-reviews.html
|   |   |   |-- professor.html
|   |   |   |-- profile.html
|   |   |   |-- project.html
|   |   |   |-- projects.html
|   |   |   `-- settings.html
|   |   |-- handlers
|   |   |   |-- admin_handler.go
|   |   |   |-- admin_handler_projects.go
|   |   |   |-- admin_handler_users.go
|   |   |   |-- auth_cookies.go
|   |   |   |-- auth_groups_handler.go
|   |   |   |-- auth_handler.go
|   |   |   |-- auth_handler_test.go
|   |   |   |-- auth_settings_handler.go
|   |   |   |-- dev_tester_handler.go
|   |   |   |-- health_handler_test.go
|   |   |   |-- health_handlers.go
|   |   |   |-- kb_handler.go
|   |   |   |-- kb_handler_test.go
|   |   |   |-- notifications_handler.go
|   |   |   |-- notify_helper.go
|   |   |   |-- project_flow_handler.go
|   |   |   |-- project_flow_handler_access.go
|   |   |   |-- project_flow_handler_grading.go
|   |   |   |-- project_flow_handler_members.go
|   |   |   |-- project_flow_handler_professors.go
|   |   |   |-- project_flow_handler_project.go
|   |   |   |-- project_flow_handler_tasks.go
|   |   |   |-- projects_handler.go
|   |   |   |-- projects_handler_create.go
|   |   |   |-- projects_handler_media.go
|   |   |   |-- projects_handler_read.go
|   |   |   |-- public_contact_handler.go
|   |   |   |-- public_contact_handler_test.go
|   |   |-- middleware
|   |   |   |-- auth.go
|   |   |   |-- auth_cookie_test.go
|   |   |   |-- rbac.go
|   |   |   |-- rbac_test.go
|   |   |   |-- request_logger.go
|   |   |   `-- request_logger_test.go
|   |   |-- dev_tester_test.go
|   |   |-- rbac_flags.go
|   |   |-- router.go
|   |   |-- router_admin.go
|   |   |-- router_auth.go
|   |   |-- router_dev.go
|   |   |-- router_notifications.go
|   |   |-- router_projectflow.go
|   |   |-- router_projects.go
|   |   |-- router_public.go
|   |   `-- routes_kb.go
|   |-- infra
|   |   |-- alerts
|   |   |   |-- health_monitor.go
|   |   |   |-- telegram.go
|   |   |   `-- telegram_test.go
|   |   |-- cache
|   |   |   `-- redis.go
|   |   |-- email
|   |   |   `-- smtp_sender.go
|   |   |-- images
|   |   |   `-- processor.go
|   |   `-- storage
|   |       `-- minio.go
|   |-- modules
|   |   |-- admin
|   |   |   `-- module.go
|   |   |-- auth
|   |   |   `-- module.go
|   |   |-- kb
|   |   |   `-- module.go
|   |   |-- notifications
|   |   |   `-- module.go
|   |   |-- projectflow
|   |   |   `-- module.go
|   |   |-- projects
|   |   |   `-- module.go
|   |   `-- rbac
|   |       `-- module.go
|   |-- repos
|   |   `-- postgres
|   |       |-- admin_repo.go
|   |       |-- auth_repo.go
|   |       |-- context_helpers.go
|   |       |-- demo_seed_integration_test.go
|   |       |-- kb_repo.go
|   |       |-- members_repo.go
|   |       |-- members_repo_integration_test.go
|   |       |-- notifications_repo.go
|   |       |-- position_repo.go
|   |       |-- position_repo_integratoin_test.go
|   |       |-- projectflow_repo.go
|   |       |-- projectflow_repo_access.go
|   |       |-- projectflow_repo_criteria.go
|   |       |-- projectflow_repo_helpers.go
|   |       |-- projectflow_repo_lifecycle.go
|   |       |-- projectflow_repo_members.go
|   |       |-- projectflow_repo_professors.go
|   |       |-- projectflow_repo_projects.go
|   |       |-- projectflow_repo_tasks.go
|   |       |-- projects_repo.go
|   |       |-- projects_repo_integration_test.go
|   |       |-- rbac_repo.go
|   |       |-- rbac_repo_integration_test.go
|   |       |-- tasks_repo.go
|   |       `-- tasks_repo_integration_test.go
|   |-- requestctx
|   |   `-- requestctx.go
|   |-- security
|   |   `-- passwords
|   |       |-- passwords.go
|   |       `-- passwords_test.go
|   |-- services
|   |   |-- admin
|   |   |   |-- service.go
|   |   |   `-- service_test.go
|   |   |-- auth
|   |   |   |-- errors.go
|   |   |   |-- rate_limiter.go
|   |   |   |-- service.go
|   |   |   `-- service_test.go
|   |   |-- kb
|   |   |   |-- service.go
|   |   |   `-- service_test.go
|   |   |-- notifications
|   |   |   |-- email_sender.go
|   |   |   |-- outbox_dispatcher.go
|   |   |   |-- service.go
|   |   |   `-- service_test.go
|   |   |-- projectflow
|   |   |   |-- ports.go
|   |   |   |-- service.go
|   |   |   |-- service_access.go
|   |   |   |-- service_grading.go
|   |   |   |-- service_member_access.go
|   |   |   |-- service_members.go
|   |   |   |-- service_professors.go
|   |   |   |-- service_tasks.go
|   |   |   `-- service_utils.go
|   |   |-- projects
|   |   |   |-- grantor.go
|   |   |   |-- repository.go
|   |   |   |-- service.go
|   |   |   |-- service_integration_test.go
|   |   |   |-- service_test.go
|   |   |   `-- view.go
|   |   `-- rbac
|   |       |-- authorizer.go
|   |       |-- cached_authorizer.go
|   |       |-- condition.go
|   |       |-- repository.go
|   |       |-- scope.go
|   |       |-- service.go
|   |       `-- service_test.go
|   `-- strutil
|       `-- truncate.go
|-- migrations
|   |-- 00001_init.sql
|   |-- 00002_rbac.sql
|   |-- 00003_projects.sql
|   |-- 00004_project_view_permission.sql
|   |-- 00005_faculties_groups.sql
|   |-- 00006_project_team.sql
|   |-- 00007_tasks.sql
|   |-- 00008_team_permissions.sql
|   |-- 00009_departments.sql
|   |-- 00010_auth_users.sql
|   |-- 00011_rbac_department_scope.sql
|   |-- 00012_seed_student_groups_predefined.sql
|   |-- 00013_project_management_extensions.sql
|   |-- 00014_seed_super_admin.sql
|   |-- 00015_multitenant_notifications_docs.sql
|   |-- 00016_professor_review_invites.sql
|   |-- 00017_member_invites.sql
|   |-- 00018_project_criteria_grading.sql
|   |-- 00019_invited_member_role.sql
|   |-- 00020_task_activity_and_submissions.sql
|   |-- 00021_backfill_task_activity.sql
|   |-- 00022_member_submit_for_grading.sql
|   |-- 00023_seed_demo_users_projects.sql
|   |-- 00024_auth_hardening.sql
|   |-- 00025_account_settings_media.sql
|   |-- 00026_group_management_department_based.sql
|   |-- 00027_profile_details.sql
|   |-- 00028_project_completed_status.sql
|   |-- 00029_seed_demo_completed_project.sql
|   |-- 00030_rbac_role_assignments_unique.sql
|   |-- 00031_rbac_tenant_admin.sql
|   |-- 00032_rbac_backfill_tenant_admin.sql
|   |-- 00033_rbac_project_security_permissions.sql
|   |-- 00034_knowledge_base.sql
|   |-- 00035_seed_rich_demo_project_flow.sql
|   |-- 00036_rbac_project_view_for_project_roles.sql
|   |-- 00037_rbac_professor_faculty_project_read.sql
|   |-- 00038_rbac_delegated_project_roles.sql
|   `-- 00039_project_retake_count.sql
|-- .env
|-- .gitignore
|-- docker-compose.yml
|-- go.mod
|-- go.sum
|-- Makefile
|-- PROJECT_TREE.md
`-- README.md
```
