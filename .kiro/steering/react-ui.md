---
inclusion: always
---

# Frontend vs Go

The **frontend (React)** is for display and frontend navigation: layout, tabs, dialogs, the sidebar tree, editor chrome, click/drag, and which screen or tab is showing.

**Go** handles everything else: Wails methods, SQLite (`loremetry-app.db`), disk files, analysis, saved reports, and other app logic. Go returns data; React renders it.

Do not put business logic, file I/O, or database work in the frontend. Do not add Vue or extra UI frameworks.
