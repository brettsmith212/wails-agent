import { useState, KeyboardEvent } from 'react'
import { Button } from '@/components/ui/button'
import { Textarea } from '@/components/ui/textarea'

type Message = {
  role: 'user' | 'assistant'
  content: string
}

function App() {
  const [messages, setMessages] = useState<Message[]>([])
  const [input, setInput] = useState('')

  const handleSubmit = () => {
    if (!input.trim()) return
    
    setMessages([...messages, { role: 'user', content: input.trim() }])
    setInput('')
    // Later we'll add assistant response logic here
  }

  const handleKeyDown = (e: KeyboardEvent<HTMLTextAreaElement>) => {
    if ((e.metaKey || e.ctrlKey) && e.key === 'Enter') {
      e.preventDefault()
      handleSubmit()
    }
  }

  return (
    <div className="flex flex-col h-screen max-h-screen p-4">
      {/* Chat history */}
      <div className="flex-1 overflow-y-auto mb-4 p-4 border rounded-lg">
        {messages.length === 0 ? (
          <div className="text-center text-muted-foreground py-8">
            Start a conversation
          </div>
        ) : (
          messages.map((message, index) => (
            <div 
              key={index} 
              className={`mb-4 p-3 rounded-lg ${message.role === 'user' 
                ? 'bg-primary text-primary-foreground ml-12' 
                : 'bg-muted mr-12'}`}
            >
              <p>{message.content}</p>
            </div>
          ))
        )}
      </div>

      {/* Input area */}
      <div className="border rounded-lg p-2 flex gap-2">
        <Textarea 
          placeholder="Type your message here..." 
          value={input} 
          onChange={(e) => setInput(e.target.value)} 
          onKeyDown={handleKeyDown}
          className="flex-1 min-h-[60px]"
        />
        <Button 
          onClick={handleSubmit}
          className="self-end"
        >
          Send
        </Button>
      </div>
    </div>
  )
}

export default App
