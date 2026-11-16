# My World Cup App — Beginner's Guide

Welcome! This folder explains **how My World Cup App works**, written for someone who is new to programming (or new to Go). Every chapter uses everyday analogies and [Mermaid](https://mermaid.js.org/) diagrams — GitHub and VS Code (with a Mermaid extension) render them as pictures.

You don't need to read the code first. Each chapter tells you which files to open when you're ready.

## Reading order

| # | Chapter | What you'll learn |
|---|---------|-------------------|
| 1 | [The big picture](01-big-picture.md) | What the app does, in one page |
| 2 | [Architecture](02-architecture.md) | How the code is split into layers, and why |
| 3 | [Where the data comes from](03-data-lifecycle.md) | Downloading, parsing, caching, and the offline fallback |
| 4 | [Life of a web request](04-request-lifecycle.md) | What happens between clicking a link and seeing a page |
| 5 | [Features tour](05-features.md) | Every page of the app and the code behind it |
| 6 | [Football rules in code](06-football-rules.md) | Points, tie-breaks, extra time, penalties, own goals |
| 7 | [Health checks, logs, metrics, and traces](07-observability.md) | How we know the app is alive and well, and the local Grafana stack |
| 8 | [Running and deploying](08-running-and-deploying.md) | Local run, Docker, Docker Compose, Kubernetes with Helm 4 |
| 9 | [Developer workflow](09-dev-workflow.md) | Tools, tests, linting, and a step-by-step "add a feature" example |
| — | [Glossary](glossary.md) | Short definitions of every technical term used here |

## The one-sentence summary

> The app downloads the World Cup 2026 data from the internet, keeps it in memory, does the football math (tables, brackets, top scorers), and shows the results as web pages — with no database at all.

## Other documents

- [`../README.md`](../README.md) — reference documentation (all Makefile targets, configuration, endpoints).
- [`../REQUIREMENTS.md`](../REQUIREMENTS.md) — what to install, and how.
- [`../CONTRIBUTING.md`](../CONTRIBUTING.md) — how to send a change.
