'use client'

import Editor from 'react-simple-code-editor'
import { highlight, languages } from 'prismjs'
import 'prismjs/components/prism-javascript'
import 'prismjs/components/prism-markup'
import 'prismjs/components/prism-markup-templating'
import 'prismjs/themes/prism.css'

interface CodeEditorProps {
  value: string
  onChange?: (value: string) => void
  language?: 'javascript' | 'html' | 'xml'
  readOnly?: boolean
}

export default function CodeEditor({ 
  value, 
  onChange, 
  language = 'html',
  readOnly = false 
}: CodeEditorProps) {
  const languageMap = {
    javascript: languages.js,
    html: languages.markup,
    xml: languages.markup,
  }

  return (
    <div className="border rounded-lg overflow-hidden bg-gray-50">
      <Editor
        value={value}
        onValueChange={onChange || (() => {})}
        highlight={code => highlight(code, languageMap[language], language)}
        padding={16}
        readOnly={readOnly}
        style={{
          fontFamily: '"Fira code", "Fira Mono", monospace',
          fontSize: 14,
          minHeight: 200,
        }}
      />
    </div>
  )
}