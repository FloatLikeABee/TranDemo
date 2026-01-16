# Transfinder Form Assistant Design - Presentation Site

This is a static HTML presentation site for the Transfinder Form Assistant project.

## Files

- **index.html** - Main landing page with project overview
- **architecture.html** - System architecture and structure
- **database.html** - Database design and schema
- **apis.html** - Core API endpoints documentation
- **complaint.html** - Complaint flow with Ground Control integration
- **form-generation.html** - Form generation process
- **student-report.html** - Student report generation
- **styles.css** - Shared CSS styling (dark gray and orange theme)

## Image Required

The `complaint.html` page references an image file:
- **ground-control-flow-config.png** - Screenshot of the Ground Control flow configuration interface

Please add this image file to the `presentation/` directory for it to display correctly.

## Viewing the Site

Simply open `index.html` in a web browser. All pages are linked via navigation.

For best results, use a local web server:
```bash
# Python 3
python -m http.server 8000

# Node.js (http-server)
npx http-server

# Then open http://localhost:8000/index.html
```

## Color Theme

- **Dark Gray:** #1a1a1a, #121212, #2a2a2a, #3a3a3a
- **Orange:** #FF8C00, #E67300, #FFA500
- **Text:** #e0e0e0, #b0b0b0, #808080

## Notes

- All pages are self-contained HTML files
- Navigation is consistent across all pages
- Responsive design for different screen sizes
- Professional presentation suitable for demonstrations

