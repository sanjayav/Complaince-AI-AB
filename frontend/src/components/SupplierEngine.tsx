const SupplierEngine = () => {
    const suppliers = [
        { id: 'SUP-0842', name: 'Global Logistics Corp', tier: 'Tier 1', risk: 'High', responseRate: '45%', status: 'At Risk', nextAction: 'Escalate' },
        { id: 'SUP-0193', name: 'Alumina Sources Ltd', tier: 'Tier 1', risk: 'Medium', responseRate: '100%', status: 'Compliant', nextAction: 'None' },
        { id: 'SUP-2041', name: 'TechPack Packaging', tier: 'Tier 2', risk: 'Low', responseRate: '85%', status: 'Pending Review', nextAction: 'Review Evidence' },
        { id: 'SUP-0522', name: 'Industrial ChemCo', tier: 'Tier 1', risk: 'High', responseRate: '12%', status: 'Non-Responsive', nextAction: 'Send Reminder' }
    ];

    return (
        <div className="content-area animate-slide-up">
            <header style={{ marginBottom: '2.5rem' }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                    <div>
                        <h1 style={{ fontSize: '2.5rem', marginBottom: '0.5rem', color: 'var(--accent-black)' }}>Supplier Engine</h1>
                        <p style={{ color: 'var(--text-secondary)', fontSize: '1.2rem', maxWidth: '800px' }}>
                            Automated outreach, data collection, and risk scoring across the group supply chain.
                        </p>
                    </div>
                    <div style={{ display: 'flex', gap: '1rem' }}>
                        <button className="btn btn-secondary">Launch Campaign</button>
                    </div>
                </div>
            </header>

            <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(240px, 1fr))', gap: '1.5rem', marginBottom: '2.5rem' }}>
                <div className="glass-panel" style={{ padding: '1.5rem' }}>
                    <h3 style={{ fontSize: '1rem', color: 'var(--text-secondary)', marginBottom: '0.5rem' }}>Active Suppliers</h3>
                    <div style={{ fontSize: '2.5rem', fontWeight: 700 }}>4,102</div>
                </div>
                <div className="glass-panel" style={{ padding: '1.5rem' }}>
                    <h3 style={{ fontSize: '1rem', color: 'var(--text-secondary)', marginBottom: '0.5rem' }}>Average Response Rate</h3>
                    <div style={{ fontSize: '2.5rem', fontWeight: 700, color: 'var(--accent-warning)' }}>68%</div>
                </div>
                <div className="glass-panel" style={{ padding: '1.5rem' }}>
                    <h3 style={{ fontSize: '1rem', color: 'var(--text-secondary)', marginBottom: '0.5rem' }}>High Risk Flagged</h3>
                    <div style={{ fontSize: '2.5rem', fontWeight: 700, color: 'var(--accent-danger)' }}>145</div>
                </div>
            </div>

            <div className="glass-panel" style={{ padding: '2rem' }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '1.5rem' }}>
                    <h2 style={{ fontSize: '1.5rem' }}>Supplier Watchlist</h2>
                    <select style={{ padding: '8px 16px', borderRadius: '8px', border: '1px solid var(--border-color)', background: 'var(--bg-primary)' }}>
                        <option>All Tiers</option>
                        <option>Tier 1</option>
                        <option>Tier 2</option>
                    </select>
                </div>
                
                <table style={{ width: '100%', borderCollapse: 'collapse', textAlign: 'left' }}>
                    <thead>
                        <tr style={{ borderBottom: '2px solid var(--border-color)', color: 'var(--text-secondary)' }}>
                            <th style={{ padding: '1rem' }}>Supplier Name</th>
                            <th style={{ padding: '1rem' }}>Tier</th>
                            <th style={{ padding: '1rem' }}>Risk Score</th>
                            <th style={{ padding: '1rem' }}>Response Rate</th>
                            <th style={{ padding: '1rem' }}>Status</th>
                            <th style={{ padding: '1rem' }}>Suggested Action</th>
                        </tr>
                    </thead>
                    <tbody>
                        {suppliers.map((sup, i) => (
                            <tr key={i} style={{ borderBottom: '1px solid var(--border-color)' }}>
                                <td style={{ padding: '1rem', fontWeight: 600 }}>{sup.name}</td>
                                <td style={{ padding: '1rem', color: 'var(--text-secondary)' }}>{sup.tier}</td>
                                <td style={{ padding: '1rem' }}>
                                    <span style={{ 
                                        color: sup.risk === 'High' ? 'var(--accent-danger)' : sup.risk === 'Medium' ? 'var(--accent-warning)' : 'var(--accent-success)',
                                        fontWeight: 600
                                    }}>
                                        {sup.risk}
                                    </span>
                                </td>
                                <td style={{ padding: '1rem' }}>
                                    <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                                        <div style={{ width: '60px', height: '6px', background: 'rgba(0,0,0,0.1)', borderRadius: '3px', overflow: 'hidden' }}>
                                            <div style={{ width: sup.responseRate, height: '100%', background: 'var(--accent-black)' }}></div>
                                        </div>
                                        <span style={{ fontSize: '0.85rem' }}>{sup.responseRate}</span>
                                    </div>
                                </td>
                                <td style={{ padding: '1rem' }}>
                                    <span className="badge" style={{ 
                                        background: sup.status === 'Compliant' ? 'rgba(34,197,94,0.1)' : sup.status === 'At Risk' || sup.status === 'Non-Responsive' ? 'rgba(239,68,68,0.1)' : 'rgba(234,179,8,0.1)',
                                        color: sup.status === 'Compliant' ? 'var(--accent-success)' : sup.status === 'At Risk' || sup.status === 'Non-Responsive' ? 'var(--accent-danger)' : 'var(--accent-warning)'
                                    }}>
                                        {sup.status}
                                    </span>
                                </td>
                                <td style={{ padding: '1rem' }}>
                                    <button style={{ 
                                        background: 'none', 
                                        border: '1px solid var(--border-color)', 
                                        padding: '4px 12px', 
                                        borderRadius: '6px',
                                        fontSize: '0.85rem',
                                        cursor: 'pointer'
                                    }}>
                                        {sup.nextAction}
                                    </button>
                                </td>
                            </tr>
                        ))}
                    </tbody>
                </table>
            </div>
        </div>
    );
};

export default SupplierEngine;
