# Frontend And Routing Review

Дата: `2026-04-23`

Scope:

- `internal/http/router_dev.go`
- `internal/http/handlers/dev_tester_handler.go`
- `internal/http/frontend/frontend.go`
- `internal/http/handlers/auth_handler.go`
- `internal/http/handlers/auth_settings_handler.go`
- `internal/http/frontend/js/auth-session.js`

## Findings

1. Medium: frontend/auth contract жёстко зашит в `/dev/*` paths.
Файлы: `internal/http/router_dev.go`, `internal/http/handlers/auth_handler.go`, `internal/http/handlers/auth_settings_handler.go`, `internal/http/frontend/js/auth-session.js`.
Детали: login redirects, verify-email redirects, password-reset redirects и client-side navigation опираются на `/dev/login`, `/dev/projects`, `/dev/admin` и смежные пути.
Риск: любое переименование URL-пространства будет дорогим и затронет handlers, JS, docs и smoke tests одновременно; кроме того, `dev` в публичном URL выглядит как след прототипа, хотя UI уже является рабочим интерфейсом.
Рекомендация: выделить единый frontend route config или constants package и перестать дублировать path strings по всему проекту.

2. Medium: регистрация встроенных страниц размазана по нескольким слоям и требует ручной синхронизации.
Файлы: `internal/http/frontend/frontend.go`, `internal/http/handlers/dev_tester_handler.go`, `internal/http/router_dev.go`, `internal/http/dev_tester_test.go`, `PROJECT_TREE.md`.
Детали: для каждой новой страницы нужно обновлять embed list, handler-функцию, router binding, тесты и документацию. Именно такой fan-out уже привёл к накоплению лишних артефактов вроде historical alias `/dev/tester`, который пришлось удалять отдельно.
Риск: drift между bundle, handlers, routes и docs; лишние маршруты и dead code будут возвращаться.
Рекомендация: перейти на table-driven registry страниц, чтобы embed/router/test wiring собирались из одного описания.

3. Low: публичная страница `/author` живёт отдельным UI-контуром и отдельными asset decisions.
Файлы: `internal/http/frontend/author.html`, `internal/http/frontend/css/author.css`, `internal/http/frontend/js/author.js`.
Детали: контактная страница использует свой layout и отдельную цепочку assets, в то время как остальной UI строится вокруг общего app-shell паттерна.
Риск: визуальные и поведенческие фиксы будут расходиться между `/author` и остальным фронтендом.
Рекомендация: если `/author` остаётся частью продукта, стоит либо подтянуть его к общему shell, либо явно оформить как isolated landing/contact surface.

## Assumptions

- Встроенный frontend остаётся частью основного продукта, а не временной demo-only оболочкой.
- `/author` сохраняется как пользовательская страница обратной связи, а не как разовый личный лендинг.
