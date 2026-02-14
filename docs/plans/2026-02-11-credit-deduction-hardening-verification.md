# Credit Deduction Hardening Verification (2026-02-11)

## Backend Data Ownership
- [x] `storyboards.user_id` startup backfill added.
- [x] Startup integrity log added: `backfilled_rows`, `remaining_zero_rows`, `mismatch_rows`, `orphan_rows`.

## Billing Path Hardening
- [x] Frame prompt: precheck required credits before creating async task.
- [x] Frame prompt: insufficient credits no longer silently fallback to free result.

## Frontend Cost Visibility
- [x] Episode workflow: show estimated total credits for:
  - Extract characters + backgrounds
  - Batch character image generation
  - Batch scene image generation
- [x] Professional editor: show estimated credits for:
  - Extract frame prompt (multiplied by call count)
  - Generate image
  - Generate video

## Frontend Balance Refresh
- [x] Refresh user credits after charge-triggering actions in:
  - `EpisodeWorkflow.vue`
  - `ProfessionalEditor.vue`

## Command Evidence
- [x] Frontend build passed:
  - `cd web; cmd /c npm run build`
- [ ] Backend `go test ./...` after this final patch set:
  - Not executable from current shell context (local `go` missing; WSL command output channel unstable in this session).

## Next Step Note (2026-02-14)
- [ ] Project desensitization:
  - Replace brand label `星亘` with `星亘` in user-facing naming and related docs/config branding text.
