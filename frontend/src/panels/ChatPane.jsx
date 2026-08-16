import { useState } from 'react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
export function ChatPane() {
    const [messages, setMessages] = useState([
        { role: 'assistant', text: 'Chat is a layout stub. No model is wired yet.' },
    ]);
    const [input, setInput] = useState('');
    function send() {
        const text = input.trim();
        if (!text)
            return;
        setInput('');
        setMessages((m) => [
            ...m,
            { role: 'user', text },
            { role: 'assistant', text: 'Noted. AI integration will answer here later.' },
        ]);
    }
    return (<div className="flex h-full flex-col bg-card">
      <div className="border-b border-border px-3 py-2 text-sm font-semibold">AI</div>
      <div className="flex-1 space-y-2 overflow-auto p-3">
        {messages.map((m, i) => (<div key={i} className={m.role === 'user'
                ? 'ml-6 rounded-md bg-primary px-3 py-2 text-sm text-primary-foreground'
                : 'mr-6 rounded-md bg-muted px-3 py-2 text-sm'}>
            {m.text}
          </div>))}
      </div>
      <form className="flex gap-2 border-t border-border p-2" onSubmit={(e) => {
            e.preventDefault();
            send();
        }}>
        <Input value={input} onChange={(e) => setInput(e.target.value)} placeholder="Ask about this story…"/>
        <Button type="submit" size="sm">
          Send
        </Button>
      </form>
    </div>);
}
