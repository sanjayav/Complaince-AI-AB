import { businessUnits } from '../data';
import type { BusinessUnit } from '../data';

const BusinessUnits = ({ onSelect }: { onSelect: (bu: BusinessUnit) => void }) => {
    return (
        <div className="content-area animate-slide-up">
            <header style={{ marginBottom: '2.5rem' }}>
                <h1 style={{ fontSize: '2.5rem', marginBottom: '0.5rem', color: 'var(--accent-black)' }}>Entity Workflows</h1>
                <p style={{ color: 'var(--text-secondary)', fontSize: '1.2rem', maxWidth: '800px' }}>
                    Manage multi-tenant governance, supplier outreach, and gap exposure across Aditya Birla business units.
                </p>
            </header>

            <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(360px, 1fr))', gap: '2rem' }}>
                {businessUnits.map((bu) => (
                    <div key={bu.id} className="glass-panel" style={{ padding: '2.5rem', display: 'flex', flexDirection: 'column', cursor: 'pointer', transition: 'all 0.3s' }} onClick={() => onSelect(bu)}>
                        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: '1.5rem' }}>
                            <div>
                                <h3 style={{ fontSize: '1.6rem', marginBottom: '0.4rem', color: 'var(--accent-black)' }}>{bu.name}</h3>
                                <span style={{ fontSize: '1rem', color: 'var(--text-secondary)' }}>{bu.sector}</span>
                            </div>
                            <span className={`badge ${bu.status === 'Compliant' ? 'badge-success' : bu.status === 'At Risk' ? 'badge-warning' : 'badge-danger'}`} style={{ fontSize: '0.85rem', padding: '8px 16px' }}>
                                {bu.status}
                            </span>
                        </div>

                        <p style={{ color: 'var(--text-secondary)', fontSize: '1.05rem', marginBottom: '2rem', lineHeight: 1.6, flex: 1 }}>
                            {bu.description}
                        </p>

                        <div style={{ display: 'flex', flexWrap: 'wrap', gap: '0.75rem', marginBottom: '2.5rem' }}>
                            {bu.modules.map(m => (
                                <span key={m} style={{ background: 'var(--bg-primary)', border: '1px solid var(--border-color)', color: 'var(--accent-black)', fontSize: '0.85rem', padding: '6px 14px', borderRadius: '100px', fontWeight: 500 }}>
                                    {m}
                                </span>
                            ))}
                        </div>

                        <button className="btn btn-secondary" style={{ width: '100%', padding: '14px', fontSize: '1rem' }}>Open Operating Layer</button>
                    </div>
                ))}
            </div>
        </div>
    );
};
export default BusinessUnits;
