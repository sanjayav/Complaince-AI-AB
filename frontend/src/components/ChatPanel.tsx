import { useState, useRef, useEffect } from 'react';

interface Message {
    id: number;
    role: 'user' | 'assistant';
    text: string;
    time: string;
}

const mockResponses: Record<string, string> = {
    'default': 'Based on our analysis of the Aditya Birla Group compliance corpus, I can provide insights on that. Could you be more specific about which business unit or framework you\'re referring to?',
    'cbam': 'The EU CBAM (Carbon Border Adjustment Mechanism) requires importers to report embedded emissions in goods like aluminium, cement, iron/steel, fertilizers, electricity, and hydrogen. For Hindalco and UltraTech, this means:\n\n1. **Scope**: All EU-bound exports of aluminium and cement\n2. **Timeline**: Full financial adjustment begins Jan 2026\n3. **Action**: Submit CBAM declarations quarterly via the EU CBAM registry\n4. **Gap**: Hindalco currently lacks verified Scope 3 data for 23% of upstream inputs',
    'brsr': 'BRSR (Business Responsibility and Sustainability Reporting) under SEBI mandates disclosure across 9 principles. Current group status:\n\n- **Hindalco**: 95% compliant, missing water stewardship KPIs\n- **UltraTech**: 88% compliant, pending Scope 3 category 15\n- **Novelis**: Not applicable (US-listed)\n- **ABF&R**: 72% compliant, needs supply chain traceability data',
    'scope': 'Scope 3 emissions across the group are tracked across 15 categories. Current coverage:\n\n- **Category 1 (Purchased goods)**: 78% coverage\n- **Category 4 (Upstream transport)**: 92% coverage\n- **Category 11 (Use of sold products)**: 45% coverage\n- **Category 15 (Investments)**: 30% coverage\n\nRecommendation: Prioritize Category 1 supplier data collection for Hindalco and UltraTech.',
    'audit': 'The Q1 2026 audit readiness across the group stands at 75%. Key findings:\n\n- **3 Critical findings** remain open (Scope 3 gaps, emission factor inconsistencies)\n- **12 actions required** across 4 business units\n- **Next milestone**: BRSR filing deadline March 31, 2026\n\nThe Audit Workspace module has the full breakdown with assignees and timelines.',
};

function getResponse(input: string): string {
    const lower = input.toLowerCase();
    if (lower.includes('cbam') || lower.includes('carbon border')) return mockResponses['cbam'];
    if (lower.includes('brsr') || lower.includes('sebi')) return mockResponses['brsr'];
    if (lower.includes('scope 3') || lower.includes('scope3') || lower.includes('emission')) return mockResponses['scope'];
    if (lower.includes('audit') || lower.includes('readiness') || lower.includes('finding')) return mockResponses['audit'];
    return mockResponses['default'];
}

function now() {
    return new Date().toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
}

const ChatPanel = ({ isOpen, onClose }: { isOpen: boolean; onClose: () => void }) => {
    const [messages, setMessages] = useState<Message[]>([
        { id: 0, role: 'assistant', text: 'Hello! I\'m your Aeiforo compliance AI assistant. Ask me about regulations, frameworks, audit readiness, or any compliance question across the Aditya Birla Group.', time: now() },
    ]);
    const [input, setInput] = useState('');
    const [typing, setTyping] = useState(false);
    const bottomRef = useRef<HTMLDivElement>(null);

    useEffect(() => {
        bottomRef.current?.scrollIntoView({ behavior: 'smooth' });
    }, [messages, typing]);

    const send = () => {
        if (!input.trim()) return;
        const userMsg: Message = { id: Date.now(), role: 'user', text: input.trim(), time: now() };
        setMessages(prev => [...prev, userMsg]);
        const q = input;
        setInput('');
        setTyping(true);
        setTimeout(() => {
            setTyping(false);
            setMessages(prev => [...prev, { id: Date.now() + 1, role: 'assistant', text: getResponse(q), time: now() }]);
        }, 1200);
    };

    if (!isOpen) return null;

    return (
        <div style={{
            position: 'fixed', bottom: '24px', right: '24px', width: '420px', height: '600px',
            background: 'rgba(255,255,255,0.92)', backdropFilter: 'blur(32px)', WebkitBackdropFilter: 'blur(32px)',
            border: '1px solid rgba(255,255,255,0.6)', borderRadius: '24px',
            boxShadow: '0 24px 48px -12px rgba(0,0,0,0.25)', display: 'flex', flexDirection: 'column',
            zIndex: 1000, overflow: 'hidden', animation: 'slideUp 0.3s cubic-bezier(0.16,1,0.3,1) forwards',
        }}>
            {/* Header */}
            <div style={{ padding: '1.25rem 1.5rem', borderBottom: '1px solid rgba(0,0,0,0.06)', display: 'flex', alignItems: 'center', justifyContent: 'space-between', flexShrink: 0 }}>
                <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem' }}>
                    <div style={{ width: '36px', height: '36px', borderRadius: '12px', background: '#0f172a', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
                        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="white" strokeWidth="2"><path d="M12 2L2 7l10 5 10-5-10-5zM2 17l10 5 10-5M2 12l10 5 10-5" /></svg>
                    </div>
                    <div>
                        <div style={{ fontWeight: 700, fontSize: '0.95rem', fontFamily: 'Outfit, sans-serif' }}>Aeiforo AI</div>
                        <div style={{ fontSize: '0.7rem', color: '#10b981', fontWeight: 600 }}>Online</div>
                    </div>
                </div>
                <button onClick={onClose} style={{ background: 'none', border: 'none', cursor: 'pointer', padding: '6px', borderRadius: '8px', color: '#6b7280' }}>
                    <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><line x1="18" y1="6" x2="6" y2="18" /><line x1="6" y1="6" x2="18" y2="18" /></svg>
                </button>
            </div>

            {/* Messages */}
            <div style={{ flex: 1, overflowY: 'auto', padding: '1rem 1.25rem', display: 'flex', flexDirection: 'column', gap: '1rem' }}>
                {messages.map(msg => (
                    <div key={msg.id} style={{ display: 'flex', flexDirection: 'column', alignItems: msg.role === 'user' ? 'flex-end' : 'flex-start' }}>
                        <div style={{
                            maxWidth: '85%', padding: '0.85rem 1rem', borderRadius: msg.role === 'user' ? '16px 16px 4px 16px' : '16px 16px 16px 4px',
                            background: msg.role === 'user' ? '#0f172a' : 'rgba(0,0,0,0.04)',
                            color: msg.role === 'user' ? 'white' : '#1e293b',
                            fontSize: '0.9rem', lineHeight: 1.6, whiteSpace: 'pre-wrap',
                        }}>
                            {msg.text}
                        </div>
                        <span style={{ fontSize: '0.65rem', color: '#9ca3af', marginTop: '4px', padding: '0 4px' }}>{msg.time}</span>
                    </div>
                ))}
                {typing && (
                    <div style={{ display: 'flex', alignItems: 'flex-start' }}>
                        <div style={{ padding: '0.85rem 1rem', borderRadius: '16px 16px 16px 4px', background: 'rgba(0,0,0,0.04)', display: 'flex', gap: '4px', alignItems: 'center' }}>
                            <div style={{ width: '6px', height: '6px', borderRadius: '50%', background: '#94a3b8', animation: 'pulse 1s infinite' }} />
                            <div style={{ width: '6px', height: '6px', borderRadius: '50%', background: '#94a3b8', animation: 'pulse 1s infinite 0.2s' }} />
                            <div style={{ width: '6px', height: '6px', borderRadius: '50%', background: '#94a3b8', animation: 'pulse 1s infinite 0.4s' }} />
                        </div>
                    </div>
                )}
                <div ref={bottomRef} />
            </div>

            {/* Quick prompts */}
            <div style={{ padding: '0 1.25rem 0.5rem', display: 'flex', gap: '0.5rem', flexWrap: 'wrap', flexShrink: 0 }}>
                {['CBAM impact?', 'BRSR status', 'Scope 3 gaps', 'Audit readiness'].map(q => (
                    <button key={q} onClick={() => { setInput(q); }} style={{ padding: '4px 10px', borderRadius: '100px', border: '1px solid rgba(0,0,0,0.08)', background: 'rgba(0,0,0,0.02)', fontSize: '0.75rem', cursor: 'pointer', color: '#6b7280', fontFamily: 'Inter, sans-serif' }}>{q}</button>
                ))}
            </div>

            {/* Input */}
            <div style={{ padding: '0.75rem 1.25rem 1.25rem', borderTop: '1px solid rgba(0,0,0,0.06)', flexShrink: 0 }}>
                <form onSubmit={e => { e.preventDefault(); send(); }} style={{ display: 'flex', gap: '0.5rem' }}>
                    <input
                        value={input} onChange={e => setInput(e.target.value)}
                        placeholder="Ask a compliance question..."
                        style={{ flex: 1, padding: '12px 16px', borderRadius: '14px', border: '1px solid rgba(0,0,0,0.08)', background: 'rgba(255,255,255,0.8)', fontSize: '0.9rem', outline: 'none', fontFamily: 'Inter, sans-serif' }}
                        onFocus={e => { e.currentTarget.style.borderColor = '#0f172a'; }}
                        onBlur={e => { e.currentTarget.style.borderColor = 'rgba(0,0,0,0.08)'; }}
                    />
                    <button type="submit" style={{ width: '44px', height: '44px', borderRadius: '14px', border: 'none', background: '#0f172a', color: 'white', cursor: 'pointer', display: 'flex', alignItems: 'center', justifyContent: 'center', flexShrink: 0 }}>
                        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><line x1="22" y1="2" x2="11" y2="13" /><polygon points="22 2 15 22 11 13 2 9 22 2" /></svg>
                    </button>
                </form>
            </div>

            <style>{`
                @keyframes pulse { 0%,100%{opacity:.3} 50%{opacity:1} }
            `}</style>
        </div>
    );
};

export default ChatPanel;
