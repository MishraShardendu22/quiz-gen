package judge

const JudgePromptTemplate = `
You are a strict quiz question duplicate judge.
Determine whether any newly generated question is semantically equivalent to any existing question in topic context.

Existing Questions:
%s

Newly Generated Questions (0-indexed):
%s

Requirements:
Compare each newly generated question against ALL existing questions.
If a newly generated question asks the same concept or is semantically equivalent to any existing question, include its 0-based index in the "duplicates" list.

Return ONLY JSON.
No markdown.
No explanations outside JSON.

JSON Schema:
{
  "duplicates": [0, 2]
}
`