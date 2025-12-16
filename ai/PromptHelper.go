package ai

import (
	"fmt"
	"strings"

	"idongivaflyinfa/models"
)

// BuildSQLPrompt constructs a prompt for SQL generation based on user request and reference SQL files
func BuildSQLPrompt(userPrompt string, sqlFiles []models.SQLFile) string {
	var contextBuilder strings.Builder
	contextBuilder.WriteString("You are a SQL expert assistant. Below are reference SQL files that you should use as examples and guidelines:\n\n")

	for _, sqlFile := range sqlFiles {
		contextBuilder.WriteString(fmt.Sprintf("--- SQL File: %s ---\n", sqlFile.Name))
		contextBuilder.WriteString(sqlFile.Content)
		contextBuilder.WriteString("\n\n")
	}

	contextBuilder.WriteString("--- User Request ---\n")
	contextBuilder.WriteString(userPrompt)
	contextBuilder.WriteString("\n\n")
	contextBuilder.WriteString("Based on the SQL files provided above, generate the correct SQL query for the user's request. Return only the SQL query without any explanation or markdown formatting.")

	return contextBuilder.String()
}

// BuildFormPrompt constructs a prompt for form JSON generation based on user request and sample JSON
func BuildFormPrompt(userPrompt string, sampleJSON string) string {
	var promptBuilder strings.Builder
	promptBuilder.WriteString("You are given a sample JSON file that represents a form entity definition for a specific web application system.\n")
	promptBuilder.WriteString("This system parses the JSON and renders it into a functional web form that users can fill out and submit.\n\n")
	promptBuilder.WriteString("The provided sample JSON represents a Student Form used to collect information from students or their parents.\n\n")
	promptBuilder.WriteString("Important Rules & Constraints:\n\n")
	promptBuilder.WriteString("The JSON structure and field names are FIXED.\n")
	promptBuilder.WriteString("You must not add, remove, rename, or restructure any fields.\n")
	promptBuilder.WriteString("The schema must remain 100%% identical to the sample.\n")
	promptBuilder.WriteString("Only the following values are allowed to change:\n")
	promptBuilder.WriteString("- Form name\n")
	promptBuilder.WriteString("- Form description\n")
	promptBuilder.WriteString("- Section name(s)\n")
	promptBuilder.WriteString("- Question titles / question names\n")
	promptBuilder.WriteString("- Other explicitly requested properties (e.g. Public flag)\n\n")
	promptBuilder.WriteString("Behavior-based rules:\n")
	promptBuilder.WriteString("- If the user requests a survey, the form must be: \"Public\": true\n")
	promptBuilder.WriteString("- If the form is for internal use, registration, or private data collection: \"Public\": false\n\n")
	promptBuilder.WriteString("All logic, field types, validation rules, and structure must remain unchanged.\n")
	promptBuilder.WriteString("You are only adapting the content, not the form mechanics.\n\n")
	promptBuilder.WriteString("Sample JSON Structure:\n")
	promptBuilder.WriteString(sampleJSON)
	promptBuilder.WriteString("\n\n--- User Request ---\n")
	promptBuilder.WriteString(userPrompt)
	promptBuilder.WriteString("\n\nBased on the user's request, generate a new form JSON that follows the exact same structure as the sample. ")
	promptBuilder.WriteString("Only modify the allowed fields (form name, description, section names, question titles/names, Public flag). ")
	promptBuilder.WriteString("Return ONLY the complete JSON object without any markdown code blocks, explanations, or additional text. ")
	promptBuilder.WriteString("The JSON must be valid and parseable.")

	return promptBuilder.String()
}

// BuildHTMLPagePrompt constructs a prompt for HTML page generation based on result file data
func BuildHTMLPagePrompt(resultFile *models.ResultFile, title string) string {
	var promptBuilder strings.Builder
	promptBuilder.WriteString("You are a professional web developer. Generate a beautiful, modern, and professional HTML page to display the following data.\n\n")

	if title != "" {
		promptBuilder.WriteString(fmt.Sprintf("Page Title: %s\n\n", title))
	}

	promptBuilder.WriteString("Data Structure:\n")
	promptBuilder.WriteString(fmt.Sprintf("Columns: %v\n", resultFile.Columns))
	promptBuilder.WriteString(fmt.Sprintf("Total Rows: %d\n\n", resultFile.RowCount))

	if resultFile.Query != "" {
		promptBuilder.WriteString(fmt.Sprintf("SQL Query: %s\n\n", resultFile.Query))
	}

	promptBuilder.WriteString("Sample Data (first 5 rows):\n")
	maxRows := 5
	if len(resultFile.Rows) < maxRows {
		maxRows = len(resultFile.Rows)
	}
	for i := 0; i < maxRows; i++ {
		promptBuilder.WriteString(fmt.Sprintf("Row %d: %v\n", i+1, resultFile.Rows[i]))
	}

	promptBuilder.WriteString("\nFull Data (all rows):\n")
	for i, row := range resultFile.Rows {
		promptBuilder.WriteString(fmt.Sprintf("Row %d: %v\n", i+1, row))
	}

	promptBuilder.WriteString("\nRequirements:\n")
	promptBuilder.WriteString("1. Create a professional, modern HTML page with a clean design\n")
	promptBuilder.WriteString("2. Use a responsive table to display ALL the data rows provided above\n")
	promptBuilder.WriteString("3. Include proper styling with CSS (embedded in <style> tag)\n")
	promptBuilder.WriteString("4. Add a header with the title\n")
	promptBuilder.WriteString("5. Show metadata section: row count, column names, timestamp\n")
	promptBuilder.WriteString("6. Make it mobile-friendly and responsive with proper table scrolling on small screens\n")
	promptBuilder.WriteString("7. Use a professional color scheme (blues, grays, whites)\n")
	promptBuilder.WriteString("8. Add hover effects on table rows for better UX\n")
	promptBuilder.WriteString("9. Include proper typography (use system fonts like -apple-system, BlinkMacSystemFont, Segoe UI)\n")
	promptBuilder.WriteString("10. Add a footer with timestamp\n")
	promptBuilder.WriteString("11. Make the table header sticky when scrolling\n")
	promptBuilder.WriteString("12. Add alternating row colors (zebra striping) for better readability\n")
	promptBuilder.WriteString("13. Include the SQL query if available in a code block\n")
	promptBuilder.WriteString("14. Add proper padding and spacing throughout\n")
	promptBuilder.WriteString("15. Use modern CSS features like flexbox/grid where appropriate\n")
	promptBuilder.WriteString("\nReturn ONLY the complete HTML code, including <!DOCTYPE html>, <html>, <head>, and <body> tags. Do not include any markdown code blocks or explanations. The HTML must be self-contained and display all rows from the data provided.")

	return promptBuilder.String()
}

// BuildFormHTMLPrompt constructs a prompt for form HTML page generation based on form JSON
func BuildFormHTMLPrompt(formJSON string, formName string, formDescription string) string {
	var promptBuilder strings.Builder
	promptBuilder.WriteString("You are a professional web developer. Generate a beautiful, modern, and professional HTML form page.\n\n")
	
	promptBuilder.WriteString("Theme Colors:\n")
	promptBuilder.WriteString("- Primary/Accent: Dark Orange (use colors like #FF8C00, #FF7F00, or #E67300)\n")
	promptBuilder.WriteString("- Background: Really Dark Grey (use colors like #1a1a1a, #1e1e1e, or #2a2a2a)\n")
	promptBuilder.WriteString("- Text: Light colors for contrast on dark background\n")
	promptBuilder.WriteString("- Use a professional dark theme with orange accents\n\n")
	
	promptBuilder.WriteString("Form Information:\n")
	if formName != "" {
		promptBuilder.WriteString(fmt.Sprintf("Form Name: %s\n", formName))
	}
	if formDescription != "" {
		promptBuilder.WriteString(fmt.Sprintf("Form Description: %s\n", formDescription))
	}
	promptBuilder.WriteString("\n")
	
	promptBuilder.WriteString("IMPORTANT: You must ONLY use the \"UDGridSections\" part of the JSON below. ")
	promptBuilder.WriteString("All other properties (InIPBoundary, RequireIPAddress, ID, DataTypeId, etc.) are configuration and should be HIDDEN from the visible form. ")
	promptBuilder.WriteString("Only render the sections and their fields (UDGridFields) as form elements.\n\n")
	
	promptBuilder.WriteString("Form JSON Structure:\n")
	promptBuilder.WriteString(formJSON)
	promptBuilder.WriteString("\n\n")
	
	promptBuilder.WriteString("Requirements:\n")
	promptBuilder.WriteString("1. Extract ONLY the UDGridSections array from the JSON\n")
	promptBuilder.WriteString("2. For each section, create a section header with the section Name\n")
	promptBuilder.WriteString("3. For each field in UDGridFields, create appropriate form inputs based on TypeName:\n")
	promptBuilder.WriteString("   - Text: <input type=\"text\">\n")
	promptBuilder.WriteString("   - Email: <input type=\"email\">\n")
	promptBuilder.WriteString("   - Phone Number: <input type=\"tel\">\n")
	promptBuilder.WriteString("   - Date/Time: <input type=\"datetime-local\">\n")
	promptBuilder.WriteString("   - Boolean: <input type=\"checkbox\"> or radio buttons\n")
	promptBuilder.WriteString("   - Currency: <input type=\"number\" step=\"0.01\">\n")
	promptBuilder.WriteString("   - Attachment: <input type=\"file\">\n")
	promptBuilder.WriteString("4. Use DisplayName for field labels\n")
	promptBuilder.WriteString("5. Mark required fields (Required: true) with an asterisk (*) and use the 'required' attribute\n")
	promptBuilder.WriteString("6. Create a professional, modern design with dark orange and really dark grey theme\n")
	promptBuilder.WriteString("7. Use proper spacing, padding, and typography\n")
	promptBuilder.WriteString("8. Make the form responsive and mobile-friendly\n")
	promptBuilder.WriteString("9. Add a submit button at the bottom\n")
	promptBuilder.WriteString("10. Include proper form validation styling\n")
	promptBuilder.WriteString("11. Use CSS embedded in <style> tag\n")
	promptBuilder.WriteString("12. Add hover effects and transitions for better UX\n")
	promptBuilder.WriteString("13. Ensure good contrast for accessibility\n")
	promptBuilder.WriteString("14. Use modern CSS features (flexbox/grid)\n")
	promptBuilder.WriteString("\nReturn ONLY the complete HTML code, including <!DOCTYPE html>, <html>, <head>, and <body> tags. ")
	promptBuilder.WriteString("Do not include any markdown code blocks or explanations. ")
	promptBuilder.WriteString("The HTML must be self-contained and render a functional form based on the UDGridSections data.")

	return promptBuilder.String()
}

