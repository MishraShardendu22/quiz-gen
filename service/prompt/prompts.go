package prompt

const DefaultAvoidInstruction = `
Do not generate questions that are semantically equivalent to any existing question.`

const AvoidDuplicatesInstruction = `

Existing Questions to Avoid:
%s
Do not generate questions that are semantically equivalent to any existing question.`

// %d is the number of questions to generate
const GeneratorPromptTemplate = `You are a quiz question generator. Generate exactly %d multiple-choice questions based ONLY on the provided context below.

Context:
%s%s

Requirements for response:
Return ONLY JSON.
No markdown.
No ` + "```json" + ` fences.
No explanations outside JSON.

The response must be a JSON object with the following schema:
{
  "questions": [
    {
      "question": "question text",
      "options": ["option 1", "option 2", "option 3", "option 4"],
      "correct_answer": 0,
      "explanation": "explanation text"
    }
  ]
}`
