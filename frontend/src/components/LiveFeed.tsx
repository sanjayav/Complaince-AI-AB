import { useState, useEffect } from 'react';

const feedItems = [
    { id: 1, time: '2 min ago', type: 'alert', source: 'EU', title: 'CBAM: Transitional period reporting deadline extended to Q2 2026', impact: 'High' },
    { id: 2, time: '18 min ago', type: 'update', source: 'India', title: 'SEBI BRSR Core: New supply chain disclosure metrics added for FY26', impact: 'Medium' },
    { id: 3, time: '34 min ago', type: 'alert', source: 'EU', title: 'CSDDD Article 22: Civil liability provisions confirmed by EU Council', impact: 'High' },
    { id: 4, time: '1 hr ago', type: 'info', source: 'Global', title: 'ISSB IFRS S2: Climate disclosure adoption tracker updated — 42 jurisdictions', impact: 'Low' },
    { id: 5, time: '1 hr ago', type: 'update', source: 'US', title: 'California SB 253: GHG reporting threshold lowered to $500M revenue', impact: 'Medium' },
    { id: 6, time: '2 hrs ago', type: 'alert', source: 'EU', title: 'EU Taxonomy: New TSC for cement sector manufacturing published', impact: 'High' },
    { id: 7, time: '3 hrs ago', type: 'info', source: 'UK', title: 'FCA: Consultation on anti-greenwashing guidance closes April 2026', impact: 'Low' },
    { id: 8, time: '4 hrs ago', type: 'update', source: 'India', title: 'MoEFCC: Updated EPR rules for packaging waste — compliance deadline Q3 2026', impact: 'Medium' },
    { id: 9, time: '5 hrs ago', type: 'alert', source: 'EU', title: 'ESRS E1: GHG emission transition plan templates released by EFRAG', impact: 'High' },
    { id: 10, time: '6 hrs ago', type: 'info', source: 'Global', title: 'GRI Universal Standards 2025 revision: Public comment period opened', impact: 'Low' },
];

const LiveFeed = () => {
    const [items, setItems] = useState(feedItems.slice(0, 5));
    const [newAlert, setNewAlert] = useState(false);

    useEffect(() => {
        const interval = setInterval(() => {
            setItems(prev => {
                const nextIdx = prev.length % feedItems.length;
                const next = feedItems[nextIdx];
                const updated = [{ ...next, id: Date.now(), time: 'Just now' }, ...prev.slice(0, 7)];
                return updated;
            });
            setNewAlert(true);
            setTimeout(() => setNewAlert(false), 2000);
        }, 8000);
        return () => clearInterval(interval);
    }, []);

    return (
        <div className="glass-panel" style={{ padding: '1.5rem', overflow: 'hidden' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1rem' }}>
                <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem' }}>
                    <h3 style={{ fontSize: '1.1rem', fontWeight: 600 }}>Live Regulatory Feed</h3>
                    <div style={{
                        width: '8px', height: '8px', borderRadius: '50%', background: '#10b981',
                        boxShadow: newAlert ? '0 0 8px 2px rgba(16,185,129,0.5)' : 'none',
                        transition: 'box-shadow 0.3s',
                    }} />
                    <span style={{ fontSize: '0.75rem', color: '#10b981', fontWeight: 600 }}>LIVE</span>
                </div>
                <span style={{ fontSize: '0.8rem', color: 'var(--text-secondary)' }}>{feedItems.length}+ sources monitored</span>
            </div>

            <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem', maxHeight: '280px', overflowY: 'auto' }}>
                {items.map((item, i) => (
                    <div key={`${item.id}-${i}`} style={{
                        display: 'flex', alignItems: 'flex-start', gap: '0.75rem', padding: '0.6rem 0.75rem',
                        borderRadius: '10px', background: i === 0 && newAlert ? 'rgba(16,185,129,0.06)' : 'transparent',
                        transition: 'background 0.5s', borderLeft: `3px solid ${item.impact === 'High' ? 'var(--accent-danger)' : item.impact === 'Medium' ? 'var(--accent-warning)' : 'rgba(0,0,0,0.1)'}`,
                    }}>
                        <div style={{ flex: 1, minWidth: 0 }}>
                            <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', marginBottom: '2px' }}>
                                <span style={{
                                    fontSize: '0.65rem', fontWeight: 700, textTransform: 'uppercase', letterSpacing: '0.05em',
                                    color: item.type === 'alert' ? 'var(--accent-danger)' : item.type === 'update' ? '#3b82f6' : 'var(--text-secondary)',
                                }}>{item.type}</span>
                                <span style={{ fontSize: '0.65rem', color: 'var(--text-secondary)' }}>•</span>
                                <span style={{ fontSize: '0.65rem', color: 'var(--text-secondary)', fontWeight: 600 }}>{item.source}</span>
                            </div>
                            <p style={{ margin: 0, fontSize: '0.85rem', lineHeight: 1.4, color: 'var(--accent-black)' }}>{item.title}</p>
                        </div>
                        <span style={{ fontSize: '0.7rem', color: 'var(--text-secondary)', whiteSpace: 'nowrap', flexShrink: 0, marginTop: '2px' }}>{item.time}</span>
                    </div>
                ))}
            </div>
        </div>
    );
};

export default LiveFeed;
