.PHONY: backend build build-prod watch typecheck

# ── Development ───────────────────────────────────────────────────────────────
# 1. (Optional) Compile TypeScript → JS whenever source changes:
#      make watch
# 2. Start the Go server:
#      make backend
# 3. Access http://localhost:8080

backend:
	cd backend && go run .

# Watch TypeScript and recompile on save (run in a separate terminal).
watch:
	cd frontend && npm run watch

# ── Production build + serve ──────────────────────────────────────────────────
# Compile TypeScript (minified) then start the Go server.
build:
	cd frontend && npm run build:prod
	cd backend && go build -o ilovmath .

run: build
	./backend/ilovmath

# Type-check only (no emit).
typecheck:
	cd frontend && npm run typecheck
