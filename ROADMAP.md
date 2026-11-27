# ROADMAP — todo-platform-eks

Đánh dấu `[x]` khi xong. Mỗi phase là 1 PR, merge xong gắn tag.

## Phase 0 — Chuẩn bị tài khoản & công cụ
- [x] AWS: MFA root, IAM user `loganlee-devops` (group `hien-admin`, MFA, không access key), account alias
- [x] AWS Budgets ($10, $20, $50) + Cost Anomaly Detection (daily email)
- [x] Tài khoản Paid plan, còn $99.99 credit (làm thêm activity để nhận đủ $200)
- [x] AWS CLI: `aws login --profile todo` (credentials tạm thời), region ap-southeast-1
- [x] Cài: git, Docker Desktop, kubectl, helm, kind, terraform, aws cli, go, node
- [x] `git init`, `.gitignore` chuẩn (template GitHub + chặn secret)
- [x] Tạo repo GitHub `lnt-h/todo-platform-eks` và push

## Phase 1 — App tối giản (`v0.1`)
- [x] Scaffold 3 app (create-vite, fastify-cli, go mod init) + `/healthz` có test
- [ ] `todo-service` (Go): CRUD todos, `/healthz`, `/metrics`, log JSON
- [ ] `user-service` (Node/TS): CRUD users, `/healthz`, `/metrics`
- [ ] `frontend` (React + Vite): danh sách todo, gọi 2 API
- [ ] Unit test cơ bản cho mỗi service

## Phase 2 — Docker (`v0.2`)
- [ ] Dockerfile multi-stage, non-root, distroless cho từng service
- [ ] `.dockerignore`
- [ ] `docker-compose.yml` (3 service + Postgres), `docker compose up` chạy được
- [ ] pre-commit + gitleaks

## Phase 3 — CI với GitHub Actions (`v0.3`)
- [ ] Workflow riêng cho từng service, có path filter
- [ ] Lint + test + build image
- [ ] Trivy scan image (fail nếu CRITICAL)
- [ ] Push image lên GHCR, tag theo git SHA
- [ ] Badge CI trên README

## Phase 4 — Kubernetes local với kind (`v0.4`)
- [ ] Helm chart: Deployment, Service, Ingress, ConfigMap, Secret
- [ ] Liveness/readiness probes, resources requests/limits
- [ ] HPA + PodDisruptionBudget
- [ ] values-dev.yaml / values-prod.yaml
- [ ] Cài thử ArgoCD trên kind

## Phase 5 — Terraform + AWS (`v0.5`) 💰 bắt đầu tốn tiền
- [ ] `infra/bootstrap`: S3 bucket state, GitHub OIDC provider + IAM role
- [ ] `infra/network`: VPC, public/private subnets, 1 NAT gateway
- [ ] `infra/eks`: EKS + managed node group Spot t3.medium
- [ ] `infra/rds`: RDS Postgres db.t4g.micro
- [ ] `infra/ecr`: repo cho từng service
- [ ] Workflow Terraform: PR → fmt/validate/tflint/checkov/plan + Infracost comment; merge → apply (manual approval)
- [ ] Workflow `destroy` (workflow_dispatch)

## Phase 6 — GitOps trên EKS (`v0.6`)
- [ ] ArgoCD app-of-apps
- [ ] Add-ons qua ArgoCD: AWS LB Controller, External Secrets, cert-manager
- [ ] CI push ECR → tự cập nhật image tag trong `gitops/` → ArgoCD sync
- [ ] Namespace `dev` và `prod`

## Phase 7 — Observability (`v0.7`)
- [ ] kube-prometheus-stack, ServiceMonitor cho các service
- [ ] Dashboard Grafana golden signals
- [ ] Alertmanager → Slack/Discord
- [ ] Loki cho logs
- [ ] OpenTelemetry + Tempo cho traces

## Phase 8 — Progressive delivery & Security (`v0.8`)
- [ ] Argo Rollouts canary 20→50→100%, auto rollback theo metric
- [ ] cosign ký image trong CI
- [ ] Kyverno: chặn image chưa ký, bắt buộc resources/non-root
- [ ] NetworkPolicy

## Phase 9 — Portfolio (`v1.0`)
- [ ] README: sơ đồ kiến trúc, luồng CI/CD, ảnh chụp màn hình
- [ ] `docs/adr/`: lý do chọn từng công cụ
- [ ] Bảng chi phí AWS thực tế
- [ ] Incident drill: làm hỏng service → alert → rollback (có ảnh/video)
- [ ] Video demo 3–5 phút
- [ ] Đưa vào CV + LinkedIn

---

## Nhật ký
| Ngày | Việc đã làm |
|---|---|
| 2026-11-27 | Khảo sát repo tham khảo, chốt tech stack, tạo CLAUDE.md + ROADMAP.md |
| 2026-11-27 | Cài helm, kind, go, pnpm, pre-commit, trivy, terraform; git init; scaffold frontend, user-service, todo-service |
| 2026-11-27 | Tạo repo public github.com/lnt-h/todo-platform-eks, push commit đầu |
| 2026-11-27 | Phase 0 xong: IAM user + MFA, budgets, anomaly detection, aws login profile `todo` |
