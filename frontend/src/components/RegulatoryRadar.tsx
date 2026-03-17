const RegulatoryRadar = () => {
    const alerts = [
        { id: '1', title: 'EU CBAM: New Reporting Format', impact: 'High', date: '2026-03-15', status: 'Action Required', description: 'Updated reporting requirements for carbon-intensive imports into the EU. Affects aluminum and cement exports.' },
        { id: '2', title: 'India BRSR: Scope Expansion', impact: 'Medium', date: '2026-03-10', status: 'In Review', description: 'SEBI proposes extending BRSR Core to top 2000 listed entities. Additional supply chain disclosures required.' },
        { id: '3', title: 'US SEC: Climate Disclosures Stayed', impact: 'Low', date: '2026-03-05', status: 'Monitoring', description: 'Eighth Circuit Court of Appeals issued an administrative stay on the SEC climate disclosure rules.' },
        { id: '4', title: 'EU CSDDD: Phased Rollout Starts', impact: 'High', date: '2026-02-28', status: 'Action Required', description: 'Corporate Sustainability Due Diligence Directive requires immediate supplier mapping for Group tier-1 companies.' }
    ];

    return (
        <div className="content-area animate-slide-up">
            <header style={{ marginBottom: '2.5rem' }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                    <div>
                        <h1 style={{ fontSize: '2.5rem', marginBottom: '0.5rem', color: 'var(--accent-black)' }}>Regulatory Radar</h1>
                        <p style={{ color: 'var(--text-secondary)', fontSize: '1.2rem', maxWidth: '800px' }}>
                            Real-time AI monitoring of global sustainability frameworks and compliance updates.
                        </p>
                    </div>
                    <button className="btn btn-primary">+ Add Framework</button>
                </div>
            </header>

            <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(240px, 1fr))', gap: '1.5rem', marginBottom: '2.5rem' }}>
                <div className="glass-panel" style={{ padding: '1.5rem' }}>
                    <h3 style={{ fontSize: '1rem', color: 'var(--text-secondary)', marginBottom: '0.5rem' }}>Active Alerts</h3>
                    <div style={{ fontSize: '2.5rem', fontWeight: 700, color: 'var(--accent-danger)' }}>12</div>
                </div>
                <div className="glass-panel" style={{ padding: '1.5rem' }}>
                    <h3 style={{ fontSize: '1rem', color: 'var(--text-secondary)', marginBottom: '0.5rem' }}>Frameworks Tracked</h3>
                    <div style={{ fontSize: '2.5rem', fontWeight: 700, color: 'var(--accent-black)' }}>48</div>
                </div>
                <div className="glass-panel" style={{ padding: '1.5rem' }}>
                    <h3 style={{ fontSize: '1rem', color: 'var(--text-secondary)', marginBottom: '0.5rem' }}>Jurisdictions</h3>
                    <div style={{ fontSize: '2.5rem', fontWeight: 700, color: 'var(--accent-black)' }}>24</div>
                </div>
            </div>

            <div className="glass-panel" style={{ padding: '2rem' }}>
                <h2 style={{ fontSize: '1.5rem', marginBottom: '1.5rem' }}>Recent AI Inferences</h2>
                
                <div style={{ display: 'flex', flexDirection: 'column', gap: '1.5rem' }}>
                    {alerts.map(alert => (
                        <div key={alert.id} style={{ display: 'flex', gap: '1.5rem', paddingBottom: '1.5rem', borderBottom: '1px solid var(--border-color)', alignItems: 'flex-start' }}>
                            <div style={{ 
                                width: '10px', 
                                height: '10px', 
                                borderRadius: '50%', 
                                marginTop: '8px',
                                background: alert.impact === 'High' ? 'var(--accent-danger)' : alert.impact === 'Medium' ? 'var(--accent-warning)' : 'var(--accent-success)' 
                            }} />
                            <div style={{ flex: 1 }}>
                                <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '0.5rem' }}>
                                    <h3 style={{ fontSize: '1.2rem', fontWeight: 600 }}>{alert.title}</h3>
                                    <span style={{ fontSize: '0.9rem', color: 'var(--text-secondary)' }}>{alert.date}</span>
                                </div>
                                <p style={{ color: 'var(--text-secondary)', marginBottom: '1rem', lineHeight: 1.5 }}>
                                    {alert.description}
                                </p>
                                <div style={{ display: 'flex', gap: '1rem', alignItems: 'center' }}>
                                    <span className="badge" style={{ 
                                        background: alert.status === 'Action Required' ? 'rgba(239,68,68,0.1)' : 'rgba(0,0,0,0.05)',
                                        color: alert.status === 'Action Required' ? 'var(--accent-danger)' : 'var(--accent-black)'
                                    }}>
                                        {alert.status}
                                    </span>
                                    <span style={{ fontSize: '0.9rem', color: 'var(--text-secondary)', fontWeight: 500 }}>
                                        Impact: {alert.impact}
                                    </span>
                                </div>
                            </div>
                            <button className="btn btn-secondary" style={{ padding: '8px 16px', fontSize: '0.9rem' }}>
                                View Gap Analysis
                            </button>
                        </div>
                    ))}
                </div>
            </div>
        </div>
    );
};

export default RegulatoryRadar;
