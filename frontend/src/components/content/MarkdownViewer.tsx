import { useState } from 'react'
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import { Check, Copy } from 'lucide-react'

interface MarkdownViewerProps {
  content: string
}

function CodeBlock({ children, className }: { children: React.ReactNode; className?: string }) {
  const [copied, setCopied] = useState(false)
  const codeString = String(children).replace(/\n$/, '')
  const language = className ? className.replace('language-', '') : 'text'

  const handleCopy = () => {
    navigator.clipboard.writeText(codeString)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  return (
    <div className="relative group my-4 rounded-xl border border-neutral-800 bg-neutral-900/90 overflow-hidden font-mono text-xs">
      <div className="flex items-center justify-between px-4 py-2 border-b border-neutral-800 bg-neutral-950/60 text-neutral-400">
        <span className="font-semibold text-neutral-300 uppercase text-[11px] tracking-wider">{language}</span>
        <button
          onClick={handleCopy}
          className="flex items-center gap-1.5 px-2 py-1 rounded bg-neutral-800/80 hover:bg-neutral-750 text-neutral-300 hover:text-white transition text-xs"
          title="Copy to clipboard"
        >
          {copied ? <Check className="h-3.5 w-3.5 text-emerald-400" /> : <Copy className="h-3.5 w-3.5" />}
          <span>{copied ? 'Copied' : 'Copy'}</span>
        </button>
      </div>
      <pre className="p-4 overflow-x-auto text-neutral-200 leading-relaxed scrollbar-thin scrollbar-thumb-neutral-700">
        <code>{children}</code>
      </pre>
    </div>
  )
}

export default function MarkdownViewer({ content }: MarkdownViewerProps) {
  return (
    <div className="prose prose-invert max-w-none prose-neutral space-y-4 text-neutral-300 leading-relaxed">
      <ReactMarkdown
        remarkPlugins={[remarkGfm]}
        components={{
          h1: ({ ...props }) => (
            <h1 className="text-2xl font-bold tracking-tight text-white border-b border-neutral-800 pb-3 mb-6" {...props} />
          ),
          h2: ({ ...props }) => (
            <h2 className="text-xl font-semibold tracking-tight text-white mt-8 mb-4 border-b border-neutral-800/60 pb-2" {...props} />
          ),
          h3: ({ ...props }) => (
            <h3 className="text-lg font-medium text-neutral-100 mt-6 mb-3" {...props} />
          ),
          p: ({ ...props }) => (
            <p className="my-3 text-neutral-300 text-sm leading-relaxed" {...props} />
          ),
          ul: ({ ...props }) => (
            <ul className="list-disc pl-6 my-3 space-y-1 text-sm text-neutral-300" {...props} />
          ),
          ol: ({ ...props }) => (
            <ol className="list-decimal pl-6 my-3 space-y-1 text-sm text-neutral-300" {...props} />
          ),
          li: ({ ...props }) => (
            <li className="my-1" {...props} />
          ),
          blockquote: ({ ...props }) => (
            <blockquote className="border-l-4 border-indigo-500 bg-neutral-900/40 px-4 py-2 my-4 rounded-r-lg italic text-neutral-300 text-sm" {...props} />
          ),
          code({ className, children, ...props }) {
            const isInline = !className
            if (isInline) {
              return (
                <code className="px-1.5 py-0.5 rounded bg-neutral-800/80 text-amber-300 font-mono text-xs border border-neutral-750" {...props}>
                  {children}
                </code>
              )
            }
            return <CodeBlock className={className}>{children}</CodeBlock>
          },
          table: ({ ...props }) => (
            <div className="overflow-x-auto my-6 rounded-lg border border-neutral-800">
              <table className="w-full text-left text-sm" {...props} />
            </div>
          ),
          thead: ({ ...props }) => (
            <thead className="bg-neutral-900/80 text-xs uppercase font-semibold text-neutral-400 border-b border-neutral-800" {...props} />
          ),
          th: ({ ...props }) => (
            <th className="px-4 py-3" {...props} />
          ),
          td: ({ ...props }) => (
            <td className="px-4 py-3 border-b border-neutral-800/60 text-neutral-300" {...props} />
          ),
        }}
      >
        {content}
      </ReactMarkdown>
    </div>
  )
}
