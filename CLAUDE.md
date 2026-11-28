# CLAUDE.md — todo-platform-eks

Dự án portfolio DevOps để chuẩn bị đi thực tập. **Trọng tâm là DevOps, không phải app.**
App Todo chỉ cần đủ đơn giản để có cái mà build, deploy, monitor.

Tiến độ chi tiết: xem [ROADMAP.md](ROADMAP.md). Sổ tay tự học: [docs/LEARNING.md](docs/LEARNING.md).

## Nguyên tắc
- App giữ tối giản (CRUD + `/healthz` + `/metrics` + log JSON). Thời gian dồn cho DevOps.
- **Không bao giờ commit secret**: `.env`, `*.key`, `*.pem`, `terraform.tfstate`, kubeconfig. Secret thật nằm ở AWS Secrets Manager → External Secrets.
- Mọi hạ tầng đều là code (Terraform / Helm / ArgoCD). Không click tay trên AWS Console (trừ bước bootstrap tài khoản).
- CI đăng nhập AWS bằng **OIDC**, không dùng access key.
- Mỗi phase = 1 branch + 1 PR, merge xong gắn tag `v0.<phase>`. Tên branch/commit/PR mô tả nội dung thay đổi (`ci/github-actions`, `build: add docker compose stack`), **không đặt theo phase**.
- Commit: Conventional Commits 1 dòng ngắn (`feat(todo-service): add crud endpoints`), không body dài. Author: `Hien Le` (GitHub `lnt-h`, email noreply, cấu hình local trong repo). Chỉ thêm trailer Co-Authored-By: Claude.
- `main` có ruleset: chỉ merge qua PR (squash), check `gitleaks` + `actionlint` phải xanh, cấm force push. Action trong workflow pin theo commit SHA.
- **Tiết kiệm credit AWS ($200)**: chỉ bật EKS khi làm việc, `terraform destroy` cuối buổi. Có Budget alert $10/$50/$100.

## Tech stack
| Lớp | Công cụ |
|---|---|
| Frontend | React + Vite, serve bằng nginx |
| todo-service | Go (net/http hoặc chi), Postgres |
| user-service | Node.js (TypeScript, Fastify), Postgres |
| Container | Docker multi-stage, distroless/non-root, docker-compose cho local |
| CI | GitHub Actions (path filter, lint, test, Trivy, gitleaks) |
| Registry | GHCR (giai đoạn đầu) → Amazon ECR |
| IaC | Terraform + terraform-aws-modules, state S3 (`use_lockfile`), tflint, checkov, Infracost |
| Cloud | AWS: VPC, EKS (Spot nodes), RDS Postgres, ECR, Secrets Manager |
| K8s packaging | Helm chart (values dev/prod) |
| CD / GitOps | ArgoCD (app-of-apps) + Argo Rollouts (canary) |
| Ingress / TLS | AWS Load Balancer Controller + cert-manager |
| Secrets | External Secrets Operator + AWS Secrets Manager |
| Observability | kube-prometheus-stack (Prometheus, Grafana, Alertmanager), Loki, OpenTelemetry + Tempo |
| Security | Trivy, gitleaks, checkov, cosign + Kyverno, NetworkPolicy |
| Local K8s | kind |

## Cấu trúc repo (mục tiêu)
```
apps/            frontend, todo-service, user-service
infra/           Terraform: bootstrap, network, eks, rds, ecr
gitops/          ArgoCD apps, Helm chart, values dev/prod
observability/   dashboards, alert rules
.github/workflows/
docs/            kiến trúc, ADR (lý do chọn công cụ), runbook
```

## Lệnh thường dùng
_(cập nhật dần khi các phase hoàn thành)_
```bash
# Test
cd apps/todo-service && go test -race ./...
cd apps/user-service && npm test
cd apps/frontend && npm run lint && npm run build

# Chạy toàn bộ bằng Docker (khuyên dùng): http://localhost:8000
docker compose up -d --build --wait
docker compose logs -f todo-service
docker compose down          # thêm -v để xóa luôn dữ liệu Postgres

# Quét lỗ hổng image
trivy image --severity HIGH,CRITICAL --ignore-unfixed todo-platform/frontend:dev

# Lint workflow GitHub Actions trước khi push
actionlint

# Xem CI
gh run list --branch <branch>
gh pr checks <số PR>

# pre-commit (chạy tự động khi commit; chạy tay cho mọi file)
pre-commit run --all-files

# Kubernetes local (kind): cài đặt đầy đủ xem deploy/kind/README.md, app ở http://localhost
helm lint gitops/charts/todo-platform -f gitops/charts/todo-platform/values-dev.yaml
helm upgrade --install todo gitops/charts/todo-platform -n todo-dev --create-namespace \
  -f gitops/charts/todo-platform/values-dev.yaml --wait
kubectl -n todo-dev get pods,svc,ingress,hpa
kind load docker-image todo-platform/<app>:dev --name todo   # sau khi build lại image

# Chạy từng service không qua Docker (cần Postgres riêng)
DATABASE_URL=postgres://... go run ./cmd/server          # todo-service :8080
DATABASE_URL=postgres://... npm run dev                  # user-service :3000
npm run dev                                              # frontend :5173, proxy /api/* tới 2 service
```
Lưu ý: máy người dùng có app khác chiếm cổng 5173, chạy Vite bằng `--port 15173` khi cần.

## Ghi chú cho Claude
- Người dùng là sinh viên đang học DevOps: giải thích ngắn gọn **tại sao** khi thêm công cụ/config mới, trả lời bằng tiếng Việt.
- Sau mỗi bước hoàn thành, tick checkbox tương ứng trong ROADMAP.md và ghi 1 dòng vào mục "Nhật ký" ở cuối file đó.
- Mỗi khi giải thích khái niệm mới, ghi vào `docs/LEARNING.md` theo phase (là gì / tại sao / áp dụng ở đâu) và thêm câu hỏi phỏng vấn cuối phase.
- Trước khi chạy lệnh tạo tài nguyên AWS tốn tiền (`terraform apply`), nhắc ước tính chi phí và nhắc destroy.
