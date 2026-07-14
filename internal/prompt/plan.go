package prompt

import "fmt"

const planModeFullReminder = `Plan mode is active. The user indicated that they do not want you to execute yet — you MUST NOT make any edits (with the exception of the plan file mentioned below), run any non-readonly tools (including changing configs or making commits), or otherwise make any changes to the system. This supercedes any other instructions you have received.

## Plan File Info:
%s
You should build your plan incrementally by writing to or editing this file. NOTE that this is the only file you are allowed to edit — other than this you are only allowed to take READ-ONLY actions.

## Plan Workflow

### Phase 1: Exploration
Goal: Gain a comprehensive understanding of the user's request by reading through code and asking them questions.

1. Focus on understanding the user's request and the code associated with their request. Actively search for existing functions, utilities, and patterns that can be reused — avoid proposing new code when suitable implementations already exist.
2. Use Glob and Grep to find relevant files and code patterns. Use Reader to read key files.
3. Launch multiple independent Glob / Grep searches in parallel where possible.

### Phase 2: Design
Goal: Design an implementation approach based on your exploration results.

Consider:
- The approach that best matches existing patterns in the codebase
- Concrete file paths and changes needed
- Potential edge cases and risks

### Phase 3: Review
Goal: Review your design and ensure alignment with the user's intentions.

1. Read the critical files to deepen your understanding
2. Ensure that the plan aligns with the user's original request
3. Use AskUserQuestion to clarify any remaining questions with the user

### Phase 4: Final Plan
Goal: Write your final plan to the plan file (the only file you can edit).
- Begin with a **Context** section: explain why this change is being made
- Include your recommended approach
- Include the paths of critical files to be modified
- Reference existing functions and utilities that should be reused, with their file paths
- Include a verification section describing how to test the changes end-to-end

### Phase 5: Completion
When you have written a thorough plan to the plan file and asked any outstanding questions, inform the user that the plan is ready for their review. Do NOT start implementing until the user explicitly asks you to proceed.

NOTE: At any point through this workflow you should feel free to ask the user questions or clarifications using the AskUserQuestion tool. Don't make large assumptions about user intent. The goal is to present a well-researched plan before implementation begins.`

const planModeSparseReminder = `Plan mode still active (see full instructions earlier in conversation). Read-only except plan file (%s). Follow the 5-phase workflow. Use AskUserQuestion for clarifications. Do NOT make any edits except to the plan file.`

const planModeExitReminder = `## Exited Plan Mode

You have exited plan mode. You can now make edits, run tools, and take actions.%s`

const reminderInterval = 5

func BuildPlanModeReminder(planFilePath string, planExists bool, iteration int) string {
	planFileInfo := fmt.Sprintf("Plan file: %s", planFilePath)
	if planExists {
		planFileInfo += "\nA plan file already exists at " + planFilePath + ". You can read it and make incremental edits using the Editor tool."
	} else {
		planFileInfo += "\nNo plan file exists yet. You should create your plan at " + planFilePath + " using the Writer tool."
	}

	if iteration == 1 {
		return fmt.Sprintf(planModeFullReminder, planFileInfo)
	}

	attachmentIndex := (iteration - 1) / reminderInterval
	if attachmentIndex%reminderInterval == 0 {
		return fmt.Sprintf(planModeFullReminder, planFileInfo)
	}

	return fmt.Sprintf(planModeSparseReminder, planFilePath)
}

func BuildPlanModeExitReminder(planFilePath string, planExists bool) string {
	extra := ""
	if planExists {
		extra = " The plan file is located at " + planFilePath + " if you need to reference it."
	}
	return fmt.Sprintf(planModeExitReminder, extra)
}
