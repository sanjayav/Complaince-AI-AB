import { businessUnits } from '../data';
import LiveFeed from './LiveFeed';

const Dashboard = () => {
    return (
        <div className="content-area animate-slide-up">
            <header style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
                <div>
                    <h1 style={{ fontSize: '2.5rem', marginBottom: '0.5rem', color: 'var(--accent-black)', letterSpacing: '-0.02em' }}>Group Compliance Intelligence</h1>
                    <p style={{ color: 'var(--text-secondary)', fontSize: '1.1rem', maxWidth: '700px' }}>
                        Aeiforo Compliance AI helps diversified industrial groups convert regulatory change into site-level actions, supplier evidence collection, and audit-ready workflows.
                    </p>
                </div>
                <div style={{ display: 'flex', gap: '1rem', alignItems: 'center' }}>
                    <button className="btn btn-secondary" style={{ borderRadius: '50%', padding: '12px', width: '48px', height: '48px' }}>
                        <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><path d="M18 8A6 6 0 0 0 6 8c0 7-3 9-3 9h18s-3-2-3-9" /><path d="M13.73 21a2 2 0 0 1-3.46 0" /></svg>
                    </button>
                    <button className="btn btn-primary">+ Create Audit Pack</button>
                </div>
            </header>

            {/* Architecture Strip */}
            <div className="glass-panel" style={{ padding: '1.5rem', display: 'flex', alignItems: 'center', justifyContent: 'space-between', fontSize: '0.85rem', fontWeight: 600, color: 'var(--accent-black)', overflowX: 'auto', whiteSpace: 'nowrap' }}>
                <span>Regulation Feeds</span>
                <span style={{ color: 'var(--text-secondary)' }}>→</span>
                <span style={{ color: 'var(--accent-success)' }}>Aeiforo AI</span>
                <span style={{ color: 'var(--text-secondary)' }}>→</span>
                <span>Group Control Library</span>
                <span style={{ color: 'var(--text-secondary)' }}>→</span>
                <span>BU Obligations</span>
                <span style={{ color: 'var(--text-secondary)' }}>→</span>
                <span>Plant/Site Tasks</span>
                <span style={{ color: 'var(--text-secondary)' }}>→</span>
                <span>Supplier Evidence</span>
                <span style={{ color: 'var(--text-secondary)' }}>→</span>
                <span style={{ background: 'var(--accent-black)', color: 'white', padding: '6px 14px', borderRadius: '100px' }}>Audit-Ready Output</span>
            </div>

            <div style={{ display: 'grid', gridTemplateColumns: '1.2fr 1fr 1fr', gap: '1.5rem' }}>
                {/* Dark Card */}
                <div className="glass-panel-dark" style={{ padding: '2rem', display: 'flex', flexDirection: 'column', justifyContent: 'space-between' }}>
                    <div>
                        <h3 style={{ fontSize: '1.1rem', fontWeight: 500, color: '#94a3b8', marginBottom: '1rem' }}>Global Exposure</h3>
                        <div style={{ display: 'flex', alignItems: 'baseline', gap: '1rem' }}>
                            <span style={{ fontSize: '4rem', fontWeight: 700, lineHeight: 1 }}>84</span>
                            <span style={{ fontSize: '1.1rem', color: '#94a3b8' }}>Obligations tracked</span>
                        </div>
                    </div>
                    <div style={{ display: 'flex', gap: '1rem', marginTop: '2.5rem' }}>
                        <div style={{ background: 'rgba(255,255,255,0.1)', padding: '1.5rem', borderRadius: '16px', flex: 1 }}>
                            <div style={{ fontSize: '2rem', fontWeight: 700, marginBottom: '0.25rem' }}>4</div>
                            <div style={{ fontSize: '0.85rem', color: '#94a3b8' }}>Business Units</div>
                        </div>
                        <div style={{ background: 'var(--accent-white)', color: 'var(--accent-black)', padding: '1.5rem', borderRadius: '16px', flex: 1 }}>
                            <div style={{ fontSize: '2rem', fontWeight: 700, marginBottom: '0.25rem' }}>12</div>
                            <div style={{ fontSize: '0.85rem' }}>Actions Required</div>
                        </div>
                    </div>
                </div>

                {/* Info Card 1 */}
                <div className="glass-panel" style={{ padding: '2.5rem' }}>
                    <h3 style={{ fontSize: '1.2rem', fontWeight: 600, marginBottom: '2rem' }}>Regulatory Action Flow</h3>
                    <ul style={{ listStyle: 'none', padding: 0, margin: 0, display: 'flex', flexDirection: 'column', gap: '1.5rem' }}>
                        <li style={{ display: 'flex', alignItems: 'center', gap: '1rem' }}>
                            <div style={{ width: '12px', height: '12px', borderRadius: '50%', background: 'var(--accent-danger)' }} />
                            <div style={{ flex: 1, fontSize: '1.1rem' }}>Reduce interpretation time</div>
                            <span style={{ fontWeight: 600, fontSize: '1.1rem' }}>-40%</span>
                        </li>
                        <li style={{ display: 'flex', alignItems: 'center', gap: '1rem' }}>
                            <div style={{ width: '12px', height: '12px', borderRadius: '50%', background: 'var(--accent-warning)' }} />
                            <div style={{ flex: 1, fontSize: '1.1rem' }}>Standardise evidence</div>
                            <span style={{ fontWeight: 600, fontSize: '1.1rem' }}>98%</span>
                        </li>
                        <li style={{ display: 'flex', alignItems: 'center', gap: '1rem' }}>
                            <div style={{ width: '12px', height: '12px', borderRadius: '50%', background: 'var(--accent-success)' }} />
                            <div style={{ flex: 1, fontSize: '1.1rem' }}>Traceable execution</div>
                            <span style={{ fontWeight: 600, fontSize: '1.1rem' }}>100%</span>
                        </li>
                    </ul>
                </div>

                {/* Info Card 2 */}
                <div className="glass-panel" style={{ padding: '2.5rem', display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', textAlign: 'center' }}>
                    <h3 style={{ fontSize: '1.2rem', fontWeight: 600, marginBottom: '0.5rem' }}>Audit Readiness</h3>
                    <p style={{ color: 'var(--text-secondary)', fontSize: '0.95rem', marginBottom: '2rem' }}>Across metals, cement, chemicals, textiles</p>

                    <div style={{ position: 'relative', width: '140px', height: '140px', display: 'flex', alignItems: 'center', justifyContent: 'center', borderRadius: '50%', background: 'conic-gradient(var(--accent-black) 0% 75%, transparent 75% 100%)', border: '8px solid rgba(0,0,0,0.05)' }}>
                        <div style={{ width: '110px', height: '110px', background: 'var(--glass-bg)', backdropFilter: 'blur(10px)', borderRadius: '50%', display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: '2rem', fontWeight: 700 }}>
                            75%
                        </div>
                    </div>

                    <button className="btn btn-secondary" style={{ marginTop: '2rem', width: '100%' }}>Download Group Report</button>
                </div>
            </div>

            <LiveFeed />

            <div>
                <h2 style={{ fontSize: '1.5rem', marginBottom: '1.5rem' }}>Business Units Monitoring</h2>
                <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(320px, 1fr))', gap: '1.5rem' }}>
                    {businessUnits.map((bu) => (
                        <div key={bu.id} className="glass-panel" style={{ padding: '1.5rem', position: 'relative', overflow: 'hidden' }}>
                            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: '1rem' }}>
                                <div>
                                    <h3 style={{ fontSize: '1.3rem', marginBottom: '0.25rem' }}>{bu.name}</h3>
                                    <span style={{ fontSize: '0.9rem', color: 'var(--text-secondary)' }}>{bu.sector}</span>
                                </div>
                                <span className={`badge ${bu.status === 'Compliant' ? 'badge-success' : bu.status === 'At Risk' ? 'badge-warning' : 'badge-danger'}`}>
                                    {bu.status}
                                </span>
                            </div>

                            <div style={{ background: 'rgba(0,0,0,0.03)', padding: '1rem', borderRadius: '12px', fontSize: '0.95rem', marginBottom: '1.5rem', minHeight: '90px' }}>
                                <strong style={{ display: 'block', marginBottom: '0.5rem', color: 'var(--text-secondary)', fontSize: '0.8rem', textTransform: 'uppercase' }}>AI Query Example</strong>
                                <span style={{ fontStyle: 'italic', color: 'var(--accent-black)' }}>"{bu.aiInsight}"</span>
                            </div>

                            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                                <span style={{ fontWeight: 600, fontSize: '1.1rem' }}>Score: {bu.score}</span>
                            </div>
                        </div>
                    ))}
                </div>
            </div>
        </div>
    );
};

export default Dashboard;
