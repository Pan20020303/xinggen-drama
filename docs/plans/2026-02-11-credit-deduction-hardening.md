# Credit Deduction Hardening Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Make credit deduction consistently correct, visible, and auditable across all AI workflows.

**Architecture:** Keep billing authoritative in backend service layer, add startup data-fix/self-check for historical ownership issues, and refresh user balance in frontend after charge-triggering actions. Validate with focused tests and end-to-end checks.

## Step 1: Backfill and Self-Check for `storyboards.user_id`
1. Add a database data-fix function to backfill `storyboards.user_id` from `episodes.user_id` where missing (`user_id=0`).
2. Add startup self-check metrics (backfilled rows, remaining zero rows, mismatch rows, orphan rows).
3. Add unit tests for backfill behavior.
4. Run backend tests.
5. Commit.

## Step 2: Charge Path Inventory and Gap Fixes
1. Enumerate all entry points that trigger LLM/image/video calls.
2. For each path, verify reserve/refund/insufficient-credit mapping.
3. Patch missing billing hooks if any.
4. Add/adjust tests for missing hooks.
5. Run backend tests.
6. Commit.

## Step 3: Frontend Total-Cost Hints for Batch Actions
1. Identify batch actions where total charge should be shown before click.
2. Show `estimated_total = unit_cost * call_count` on action buttons/dialogs.
3. Handle zero/unconfigured price fallback in UI.
4. Build frontend.
5. Commit.

## Step 4: Balance Refresh Coverage
1. Ensure all charge-triggering frontend actions refresh `/auth/me` after task creation.
2. Keep refresh non-blocking for primary action success/failure messaging.
3. Build frontend and run backend tests.
4. Commit.

## Step 5: Verification Checklist
1. Validate text/image/video deduction with positive costs.
2. Validate insufficient credits returns business error (403) and no background task leakage.
3. Validate failed generation leads to refund records.
4. Commit any checklist/doc updates.

