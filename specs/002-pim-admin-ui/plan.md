# Implementation Plan: PIM Admin UI

**Branch**: `002-pim-admin-ui` | **Date**: 2025-11-14 | **Spec**: [spec.md](./spec.md)

## Summary

Build a web-based admin interface for the PIM system that allows merchandising managers to manage products and variants through a clean, functional UI. The interface is a single-page application using vanilla JavaScript that calls the existing REST API endpoints. No backend changes required - pure frontend addition served as static HTML/CSS/JS from the Go server.

## Technical Context

**Language/Version**: HTML5, CSS3, vanilla JavaScript (ES6+)  
**Primary Dependencies**: None (no npm, no frameworks, no build step)  
**API Integration**: Fetch API calling existing /api/v1/products and /api/v1/variants endpoints  
**Serving**: Static HTML served from Go server at /admin route  
**Storage**: No additional storage (uses existing PostgreSQL via API)  
**Testing**: Manual browser testing (no automated UI tests for MVP)  
**Target Platform**: Desktop browsers (Chrome, Firefox, Safari, Edge)  
**Project Type**: Single HTML page with embedded CSS and JavaScript  
**Performance Goals**: Page load <2s, API operations <1s response time  
**Constraints**: No authentication, desktop-first, calls existing API only  
**Scale/Scope**: Manage existing product catalog, no new backend features

## Constitution Check

This feature is **UI-only** and doesn't require backend code changes or tests per constitution:
- ⚠️ **Integration Testing**: Not applicable (frontend only, no Go code)
- ⚠️ **Table-Driven Tests**: Not applicable (frontend only)
- ⚠️ **Edge Case Coverage**: Manual browser testing sufficient for UI
- ⚠️ **Real Database Fixtures**: Uses existing API (database via API only)
- ⚠️ **ServeHTTP Testing**: Not applicable (static HTML serving)

**Justification**: This is a pure frontend feature that adds a static HTML interface. All business logic remains in the existing tested API. The UI is a thin client that makes API calls.

## Project Structure

```text
/Users/sunfmin/Developments/apidemo1/
├── web/
│   └── admin.html          # Single-page admin interface (HTML+CSS+JS)
├── internal/handlers/
│   └── admin.go            # Handler to serve static HTML
└── cmd/api/main.go         # Updated to mount /admin route
```

**Structure Decision**: Single HTML file approach for simplicity. All HTML, CSS, and JavaScript in one file for easy deployment. Served as static content from Go server.

## Implementation Tasks

### Phase 1: Static HTML Setup (2 tasks)
1. Create admin.html with HTML structure, embedded CSS, and embedded JavaScript
2. Create handler in internal/handlers/admin.go to serve the HTML file
3. Mount /admin route in cmd/api/main.go

### Phase 2: Product Management UI (3 tasks)
1. Implement product list view with API integration
2. Implement product create/edit form with dynamic attributes
3. Implement product delete functionality

### Phase 3: Variant Management UI (2 tasks)
1. Implement variant list view within product details
2. Implement variant create/edit/delete functionality

**Total**: 7 implementation tasks (no tests per constitution exemption for UI-only features)

## Quick Implementation

Since this is a straightforward UI addition, we can implement it directly without extensive planning artifacts. The interface will:
- Use Fetch API to call existing endpoints
- Display data in clean tables
- Provide forms for CRUD operations
- Handle API errors gracefully
- Use modern CSS for styling
