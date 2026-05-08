# Contributing

This repository is an Overture dogfood target. Changes are authored by
agent teams, coordinated through Band rooms, and gated by human approval.

Use this file for review behavior on Overture-authored PRs. It does not
define a general open-source contribution process.

## Human review feedback

Leave feedback as GitHub PR review comments so the agent team can track,
answer, and resolve it in the same place as the diff.

When feedback applies to a specific file or line, prefer an inline review
comment on that line. Use a top-level PR comment only when the issue spans
multiple files, the requested behavior is cross-cutting, or the question is
about the PR as a whole.

State the expected behavior. Prefer "return `ErrInvalidConfig` when the
token is empty" over "make config validation better".

Separate blocking feedback from suggestions. Blocking feedback should cover
correctness, security, data loss, API compatibility, license, or repository
policy issues that must be fixed before approval. Suggestions should describe
style, naming, wording, or local cleanup that would improve the change but
does not block the PR.

Keep feedback actionable. If a comment asks for a different design, name the
constraint the current design violates or the case it fails to support.
Reviewers should leave specific, actionable feedback rather than vague reactions such as "looks weird".

## Agent response etiquette

Acknowledge review comments without arguing. If the reviewer is correct, make
the change. If the comment is ambiguous, ask one narrow clarifying question in
the PR thread before editing.

Make the smallest repo-consistent change that satisfies the review. Do not
expand the PR into nearby cleanup, architecture work, or unrelated fixes.

Preserve human edits and unrelated work already present in the branch. Read
the current diff before editing files touched by another person or agent.

When a requested change cannot be made safely, explain the specific blocker in
the PR thread and stop editing that area until the reviewer clarifies the path
forward.

After addressing review, report what changed and include validation evidence.
Evidence can be a command output, a diff reference, or a short note explaining
why no code validation was needed for a documentation-only change.
