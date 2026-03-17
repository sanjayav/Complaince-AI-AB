const AuditWorkspace = () => {
    const audits = [
        { id: 'AUD-2026-Q1', title: 'Q1 Corporate Sustainability Audit', framework: 'BRSR', status: 'In Progress', progress: '75%', dueDate: '2026-03-31', assignee: 'Sarah Jenkins' },
        { id: 'AUD-2026-EU', title: 'CBAM Readiness Review', framework: 'EU CBAM', status: 'Planning', progress: '10%', dueDate: '2026-04-15', assignee: 'Michael Chang' },
        { id: 'AUD-2025-YE', title: 'FY25 Annual GHG Inventory', framework: 'GHG Protocol', status: 'Completed', progress: '100%', dueDate: '2026-01-31', assignee: 'Priya Sharma' }
    ];

    const findings = [
        { id: 'FND-041', severity: 'Critical', description: 'Missing Scope 3 category 1 data for top 50 suppliers.', audit: 'Q1 Corporate Sustainability Audit' },
        { id: 'FND-042', severity: 'Medium', description: 'Inconsistent emission factors used across EU plants.', audit: 'CBAM Readiness Review' },
        { id: 'FND-043', severity: 'Low', description: 'Water usage meter calibration logs outdated for Plant B.', audit: 'Q1 Corporate Sustainability Audit' }
    ];

    return (
        <div className="content-area animate-slide-up">
            <header style={{ marginBottom: '2.5rem' }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                    <div>
                        <h1 style={{ fontSize: '2.5rem', marginBottom: '0.5rem', color: 'var(--accent-black)' }}>Audit Workspace</h1>
                        <p style={{ color: 'var(--text-secondary)', fontSize: '1.2rem', maxWidth: '800px' }}>
                            Generate audit-ready evidence packs and track internal compliance readiness.
                        </p>
                    </div>
                    <button className="btn btn-primary">+ New Audit Plan</button>
                </div>
            </header>

            <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(300px, 1fr))', gap: '1.5rem', marginBottom: '2.5rem' }}>
                {audits.map((audit) => (
                    <div key={audit.id} className="glass-panel" style={{ padding: '1.5rem' }}>
                        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: '1rem' }}>
                            <h3 style={{ fontSize: '1.1rem', fontWeight: 600 }}>{audit.title}</h3>
                            <span className="badge" style={{ 
                                background: audit.status === 'Completed' ? 'rgba(34,197,94,0.1)' : audit.status === 'In Progress' ? 'rgba(59,130,246,0.1)' : 'rgba(0,0,0,0.05)',
                                color: audit.status === 'Completed' ? 'var(--accent-success)' : audit.status === 'In Progress' ? '#3b82f6' : 'var(--text-secondary)'
                            }}>
                                {audit.status}
                            </span>
                        </div>
                        <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: '0.9rem', color: 'var(--text-secondary)', marginBottom: '1rem' }}>
                            <span>{audit.framework}</span>
                            <span>Due: {audit.dueDate}</span>
                        </div>
                        <div style={{ marginBottom: '1rem' }}>
                            <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: '0.85rem', marginBottom: '0.25rem' }}>
                                <span>Readiness</span>
                                <span>{audit.progress}</span>
                            </div>
                            <div style={{ width: '100%', height: '6px', background: 'rgba(0,0,0,0.1)', borderRadius: '3px', overflow: 'hidden' }}>
                                <div style={{ width: audit.progress, height: '100%', background: audit.progress === '100%' ? 'var(--accent-success)' : 'var(--accent-black)' }}></div>
                            </div>
                        </div>
                        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                            <span style={{ fontSize: '0.85rem', color: 'var(--text-secondary)' }}>Lead: {audit.assignee}</span>
                            <button style={{ background: 'none', border: 'none', color: 'var(--accent-black)', fontWeight: 600, fontSize: '0.9rem', cursor: 'pointer' }}>Open →</button>
                        </div>
                    </div>
                ))}
            </div>

            <div className="glass-panel" style={{ padding: '2rem' }}>
                <h2 style={{ fontSize: '1.5rem', marginBottom: '1.5rem' }}>Open Findings & Tasks</h2>
                
                <div style={{ display: 'flex', flexDirection: 'column', gap: '1rem' }}>
                    {findings.map(finding => (
                        <div key={finding.id} style={{ display: 'flex', gap: '1.5rem', padding: '1rem', border: '1px solid var(--border-color)', borderRadius: '12px', alignItems: 'center' }}>
                            <div style={{ 
                                width: '4px', 
                                height: '40px', 
                                borderRadius: '2px', 
                                background: finding.severity === 'Critical' ? 'var(--accent-danger)' : finding.severity === 'Medium' ? 'var(--accent-warning)' : 'var(--accent-success)' 
                            }} />
                            <div style={{ flex: 1 }}>
                                <p style={{ fontWeight: 500, marginBottom: '0.25rem' }}>{finding.description}</p>
                                <div style={{ display: 'flex', gap: '1rem', fontSize: '0.85rem', color: 'var(--text-secondary)' }}>
                                    <span>{finding.id}</span>
                                    <span>•</span>
                                    <span>{finding.audit}</span>
                                </div>
                            </div>
                            <button className="btn btn-secondary" style={{ padding: '6px 12px', fontSize: '0.85rem' }}>Resolve</button>
                        </div>
                    ))}
                </div>
            </div>
        </div>
    );
};

export default AuditWorkspace;
