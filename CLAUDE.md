# CLAUDE.md — todo-platform-eks

Dự án portfolio DevOps để chuẩn bị đi thực tập. **Trọng tâm là DevOps, không phải app.**
App Todo chỉ cần đủ đơn giản để có cái mà build, deploy, monitor.

Tiến độ chi tiết: xem [ROADMAP.md](ROADMAP.md).

## Nguyên tắc
- App giữ tối giản (CRUD + `/healthz` + `/metrics` + log JSON). Thời gian dồn cho DevOps.
- **Không bao giờ commit secret**: `.env`, `*.key`, `*.pem`, `terraform.tfstate`, kubeconfig. Secret thật nằm ở AWS Secrets Manager → External Secrets.
- Mọi hạ tầng đều là code (Terraform / Helm / ArgoCD). Không click tay trên AWS Console (trừ bước bootstrap tài khoản).
- CI đăng nhập AWS bằng **OIDC**, không dùng access key.
- Mỗi phase = 1 branch + 1 PR, merge xong gắn tag `v0.<phase>`.
- Commit: Conventional Commits 1 dòng ngắn (`feat(todo-service): add crud endpoints`), không body dài. Giữ trailer Co-Authored-By: Claude.
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
docker compose up --build        # chạy local toàn bộ hệ thống
```

## Ghi chú cho Claude
- Người dùng là sinh viên đang học DevOps: giải thích ngắn gọn **tại sao** khi thêm công cụ/config mới, trả lời bằng tiếng Việt.
- Sau mỗi bước hoàn thành, tick checkbox tương ứng trong ROADMAP.md và ghi 1 dòng vào mục "Nhật ký" ở cuối file đó.
- Trước khi chạy lệnh tạo tài nguyên AWS tốn tiền (`terraform apply`), nhắc ước tính chi phí và nhắc destroy.
