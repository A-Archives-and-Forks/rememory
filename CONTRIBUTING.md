# Contributing to ReMemory

ReMemory is a project where quality matters more than usual. A recovery bundle might sit in a drawer for ten years, then be opened by someone who just lost a loved one. The code, the copy, the design — it all has to hold up. Contributions should reflect that care.

We welcome contributions. Here's how to make them count.

## Code of conduct

We follow the [Contributor Covenant](https://www.contributor-covenant.org/version/2/1/code_of_conduct/). Be kind, be respectful, don't be an asshole. If someone is making the community worse, we'll act on it.

## How to contribute

- **Small PRs, incremental improvements.** A series of focused, reviewable PRs is better than one large change.
- **Discuss before building big things.** For major features, refactors, or architectural changes, open an issue to discuss the approach first. Once there's agreement, break the work into smaller pieces and land them incrementally. Don't open a large PR out of the blue.
- **Bug fixes and small improvements can go straight to PR.** Not everything needs a discussion. If the change is self-contained and obvious, open the PR.
- **Read the code before changing it.** Understand the existing patterns, match the style, and search for prior art before introducing something new.
- **Run the tests.** `make test` for Go tests, `make test-e2e` for Playwright browser tests. Don't leave this for reviewers to discover.
- **Read `AGENTS.md`.** It contains the architectural context, development principles, and voice guidelines. Especially the voice guidelines — they matter here.

## On using AI tools

AI tools are fine. Use whatever helps you write better code. We don't care how you got to the solution — we care whether the solution is good.

The standard is the same regardless of how the code was written: **you are responsible for every line of your contribution.** This applies whether you wrote it by hand, used an AI assistant, or copied it from somewhere. If you can't explain why a line is there and why it's correct, it shouldn't be in your PR.

What we're addressing here isn't AI — it's contributions where nobody took the time to make them good. This applies equally to AI-generated and human-written slop.

**Specifically:**

- If you use AI to help write code, you must understand every line of the diff you're submitting. You must be able to explain why the change is correct and why this approach is the right one.
- Do not paste raw AI output into issues, PRs, or comments. If you use AI to help draft text, rewrite it in your own words. Make sure it's concise, relevant, and adds value. Boilerplate, filler, and formulaic responses waste everyone's time.
- Remove AI-generated footers, co-author attributions, and "Generated with..." signatures before submitting. Their presence tells us you didn't review your own submission carefully enough to notice them.
- Automated submissions — bots or agents posting PRs, comments, or issue responses without meaningful human review — will be treated as spam. The human must review, edit, and take ownership of the final submission.

We're a small project. Every PR we review takes time away from building something that matters. Submitting work you didn't review is not helping — it's creating work.

## Consequences

We'd rather help you improve a PR than close it. But we can't review work that wasn't reviewed by the person submitting it.

- **First time:** You'll get a warning and a chance to fix the PR.
- **Second time:** The PR will be closed.
- **Repeated offenses:** May be reported to GitHub as spam.

## What a good contribution looks like

A good PR is focused, tested, and matches existing patterns. The description makes clear the author understood what they were changing and why — not a wall of AI-generated text, but a concise explanation in their own words.

A small, thoughtful PR from someone who read the code is worth more than a large one from someone who didn't.
