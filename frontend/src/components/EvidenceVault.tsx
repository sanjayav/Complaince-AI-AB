const EvidenceVault = () => {
    const documents = [
        { id: 'DOC-1029', name: 'Hindalco_Q3_Emissions_Report.pdf', type: 'Report', date: '2026-03-12', bu: 'Hindalco', status: 'Verified', aiConfidence: '98%' },
        { id: 'DOC-1030', name: 'Supplier_Code_of_Conduct_2026.docx', type: 'Policy', date: '2026-03-10', bu: 'Group Wide', status: 'Draft', aiConfidence: 'N/A' },
        { id: 'DOC-1031', name: 'UltraTech_Water_Usage_Log.xlsx', type: 'Data Log', date: '2026-03-08', bu: 'UltraTech Cement', status: 'Flagged', aiConfidence: '64%' },
        { id: 'DOC-1032', name: 'Novelis_Recycling_Certificates_Batch4.zip', type: 'Evidence', date: '2026-03-05', bu: 'Novelis', status: 'Verified', aiConfidence: '95%' }
    ];

    return (
        <div className="content-area animate-slide-up">
            <header style={{ marginBottom: '2.5rem' }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                    <div>
                        <h1 style={{ fontSize: '2.5rem', marginBottom: '0.5rem', color: 'var(--accent-black)' }}>Evidence Vault</h1>
                        <p style={{ color: 'var(--text-secondary)', fontSize: '1.2rem', maxWidth: '800px' }}>
                            Centralized immutable storage for compliance documentation and automatically extracted metrics.
                        </p>
                    </div>
                    <div style={{ display: 'flex', gap: '1rem' }}>
                        <button className="btn btn-secondary">Bulk Upload</button>
                        <button className="btn btn-primary">Upload Document</button>
                    </div>
                </div>
            </header>

            <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(240px, 1fr))', gap: '1.5rem', marginBottom: '2.5rem' }}>
                <div className="glass-panel" style={{ padding: '1.5rem' }}>
                    <h3 style={{ fontSize: '1rem', color: 'var(--text-secondary)', marginBottom: '0.5rem' }}>Total Documents</h3>
                    <div style={{ fontSize: '2.5rem', fontWeight: 700 }}>45,291</div>
                </div>
                <div className="glass-panel" style={{ padding: '1.5rem' }}>
                    <h3 style={{ fontSize: '1rem', color: 'var(--text-secondary)', marginBottom: '0.5rem' }}>Data Points Extracted</h3>
                    <div style={{ fontSize: '2.5rem', fontWeight: 700, color: 'var(--accent-success)' }}>1.2M+</div>
                </div>
                <div className="glass-panel" style={{ padding: '1.5rem' }}>
                    <h3 style={{ fontSize: '1rem', color: 'var(--text-secondary)', marginBottom: '0.5rem' }}>Anomalies Flagged</h3>
                    <div style={{ fontSize: '2.5rem', fontWeight: 700, color: 'var(--accent-danger)' }}>32</div>
                </div>
            </div>

            <div className="glass-panel" style={{ padding: '2rem' }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '1.5rem' }}>
                    <h2 style={{ fontSize: '1.5rem' }}>Recent Ingestions</h2>
                    <input 
                        type="text" 
                        placeholder="Search evidence..." 
                        style={{ padding: '8px 16px', borderRadius: '8px', border: '1px solid var(--border-color)', background: 'var(--bg-primary)', width: '250px' }}
                    />
                </div>
                
                <table style={{ width: '100%', borderCollapse: 'collapse', textAlign: 'left' }}>
                    <thead>
                        <tr style={{ borderBottom: '2px solid var(--border-color)', color: 'var(--text-secondary)' }}>
                            <th style={{ padding: '1rem' }}>Document ID</th>
                            <th style={{ padding: '1rem' }}>Name</th>
                            <th style={{ padding: '1rem' }}>Type</th>
                            <th style={{ padding: '1rem' }}>Business Unit</th>
                            <th style={{ padding: '1rem' }}>Date Added</th>
                            <th style={{ padding: '1rem' }}>AI Confidence</th>
                            <th style={{ padding: '1rem' }}>Status</th>
                            <th style={{ padding: '1rem' }}></th>
                        </tr>
                    </thead>
                    <tbody>
                        {documents.map((doc, i) => (
                            <tr key={i} style={{ borderBottom: '1px solid var(--border-color)' }}>
                                <td style={{ padding: '1rem', fontSize: '0.9rem', color: 'var(--text-secondary)' }}>{doc.id}</td>
                                <td style={{ padding: '1rem', fontWeight: 500, color: 'var(--accent-black)' }}>{doc.name}</td>
                                <td style={{ padding: '1rem' }}>{doc.type}</td>
                                <td style={{ padding: '1rem' }}>{doc.bu}</td>
                                <td style={{ padding: '1rem' }}>{doc.date}</td>
                                <td style={{ padding: '1rem' }}>
                                    <span style={{ color: doc.aiConfidence === '98%' || doc.aiConfidence === '95%' ? 'var(--accent-success)' : doc.aiConfidence === '64%' ? 'var(--accent-danger)' : 'var(--text-secondary)' }}>
                                        {doc.aiConfidence}
                                    </span>
                                </td>
                                <td style={{ padding: '1rem' }}>
                                    <span className="badge" style={{ 
                                        background: doc.status === 'Verified' ? 'rgba(34,197,94,0.1)' : doc.status === 'Flagged' ? 'rgba(239,68,68,0.1)' : 'rgba(0,0,0,0.05)',
                                        color: doc.status === 'Verified' ? 'var(--accent-success)' : doc.status === 'Flagged' ? 'var(--accent-danger)' : 'var(--accent-black)'
                                    }}>
                                        {doc.status}
                                    </span>
                                </td>
                                <td style={{ padding: '1rem', textAlign: 'right' }}>
                                    <button style={{ background: 'none', border: 'none', cursor: 'pointer', color: 'var(--text-secondary)' }}>•••</button>
                                </td>
                            </tr>
                        ))}
                    </tbody>
                </table>
            </div>
        </div>
    );
};

export default EvidenceVault;
