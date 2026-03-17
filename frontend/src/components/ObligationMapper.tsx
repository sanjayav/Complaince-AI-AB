const ObligationMapper = () => {
    const obligations = [
        { id: 'OBL-001', framework: 'GRI', requirement: 'GRI 305: Emissions 2016', mapping: 'Direct (Scope 1) GHG emissions', bu: 'Hindalco', status: 'Mapped', coverage: '100%' },
        { id: 'OBL-002', framework: 'BRSR', requirement: 'Principle 6', mapping: 'Energy and GHG Emissions', bu: 'Novelis', status: 'Mapped', coverage: '95%' },
        { id: 'OBL-003', framework: 'CSRD', requirement: 'Article 15', mapping: 'Adverse environmental impacts', bu: 'Aditya Birla Fashion', status: 'Partial', coverage: '60%' },
        { id: 'OBL-004', framework: 'ISSB', requirement: 'ESRS E1', mapping: 'Climate change mitigation', bu: 'UltraTech Cement', status: 'Unmapped', coverage: '0%' }
    ];

    return (
        <div className="content-area animate-slide-up">
            <header style={{ marginBottom: '2.5rem' }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                    <div>
                        <h1 style={{ fontSize: '2.5rem', marginBottom: '0.5rem', color: 'var(--accent-black)' }}>Obligation Mapper</h1>
                        <p style={{ color: 'var(--text-secondary)', fontSize: '1.2rem', maxWidth: '800px' }}>
                            Automatically map regulatory requirements to internal business unit controls and data points.
                        </p>
                    </div>
                    <button className="btn btn-primary">Run AI Mapping</button>
                </div>
            </header>

            <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(240px, 1fr))', gap: '1.5rem', marginBottom: '2.5rem' }}>
                <div className="glass-panel" style={{ padding: '1.5rem', background: 'var(--accent-black)', color: 'white' }}>
                    <h3 style={{ fontSize: '1rem', color: '#94a3b8', marginBottom: '0.5rem' }}>Total Requirements</h3>
                    <div style={{ fontSize: '2.5rem', fontWeight: 700 }}>1,240</div>
                </div>
                <div className="glass-panel" style={{ padding: '1.5rem' }}>
                    <h3 style={{ fontSize: '1rem', color: 'var(--text-secondary)', marginBottom: '0.5rem' }}>Mapped (Auto)</h3>
                    <div style={{ fontSize: '2.5rem', fontWeight: 700, color: 'var(--accent-success)' }}>86%</div>
                </div>
                <div className="glass-panel" style={{ padding: '1.5rem' }}>
                    <h3 style={{ fontSize: '1rem', color: 'var(--text-secondary)', marginBottom: '0.5rem' }}>Pending Review</h3>
                    <div style={{ fontSize: '2.5rem', fontWeight: 700, color: 'var(--accent-warning)' }}>174</div>
                </div>
            </div>

            <div className="glass-panel" style={{ padding: '2rem' }}>
                <h2 style={{ fontSize: '1.5rem', marginBottom: '1.5rem' }}>Cross-Framework Matrix</h2>
                
                <table style={{ width: '100%', borderCollapse: 'collapse', textAlign: 'left' }}>
                    <thead>
                        <tr style={{ borderBottom: '2px solid var(--border-color)', color: 'var(--text-secondary)' }}>
                            <th style={{ padding: '1rem' }}>ID</th>
                            <th style={{ padding: '1rem' }}>Framework</th>
                            <th style={{ padding: '1rem' }}>Requirement</th>
                            <th style={{ padding: '1rem' }}>Internal Mapping</th>
                            <th style={{ padding: '1rem' }}>Business Unit</th>
                            <th style={{ padding: '1rem' }}>Coverage</th>
                            <th style={{ padding: '1rem' }}>Status</th>
                        </tr>
                    </thead>
                    <tbody>
                        {obligations.map((obl, i) => (
                            <tr key={i} style={{ borderBottom: '1px solid var(--border-color)' }}>
                                <td style={{ padding: '1rem', fontWeight: 600 }}>{obl.id}</td>
                                <td style={{ padding: '1rem' }}>
                                    <span style={{ background: 'var(--bg-primary)', padding: '4px 8px', borderRadius: '4px', fontSize: '0.85rem' }}>
                                        {obl.framework}
                                    </span>
                                </td>
                                <td style={{ padding: '1rem' }}>{obl.requirement}</td>
                                <td style={{ padding: '1rem' }}>{obl.mapping}</td>
                                <td style={{ padding: '1rem', color: 'var(--text-secondary)' }}>{obl.bu}</td>
                                <td style={{ padding: '1rem' }}>
                                    <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                                        <div style={{ width: '60px', height: '6px', background: 'rgba(0,0,0,0.1)', borderRadius: '3px', overflow: 'hidden' }}>
                                            <div style={{ width: obl.coverage, height: '100%', background: obl.coverage === '100%' ? 'var(--accent-success)' : obl.coverage === '0%' ? 'var(--accent-danger)' : 'var(--accent-warning)' }}></div>
                                        </div>
                                        <span style={{ fontSize: '0.85rem' }}>{obl.coverage}</span>
                                    </div>
                                </td>
                                <td style={{ padding: '1rem' }}>
                                    <span className="badge" style={{ 
                                        background: obl.status === 'Mapped' ? 'rgba(34,197,94,0.1)' : obl.status === 'Partial' ? 'rgba(234,179,8,0.1)' : 'rgba(239,68,68,0.1)',
                                        color: obl.status === 'Mapped' ? 'var(--accent-success)' : obl.status === 'Partial' ? 'var(--accent-warning)' : 'var(--accent-danger)'
                                    }}>
                                        {obl.status}
                                    </span>
                                </td>
                            </tr>
                        ))}
                    </tbody>
                </table>
            </div>
        </div>
    );
};

export default ObligationMapper;
