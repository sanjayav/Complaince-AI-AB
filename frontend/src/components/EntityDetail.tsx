import { useState } from 'react';
import type { BusinessUnit } from '../data';

const EntityDetail = ({ entity, onBack }: { entity: BusinessUnit; onBack: () => void }) => {
    const [activeTab, setActiveTab] = useState('Obligation Mapper');
    const tabs = ['Obligation Mapper', 'Evidence Vault', 'Supplier Outreach Engine', 'Audit Workspace'];

    return (
        <div className="content-area animate-slide-up">
            <header style={{ marginBottom: '2.5rem', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                <div style={{ display: 'flex', alignItems: 'center', gap: '2rem' }}>
                    <button className="btn btn-secondary" style={{ padding: '16px', borderRadius: '50%', width: '56px', height: '56px', display: 'flex', alignItems: 'center', justifyContent: 'center' }} onClick={onBack}>
                        <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><path d="M19 12H5M12 19l-7-7 7-7" /></svg>
                    </button>
                    <div>
                        <h1 style={{ fontSize: '2.4rem', marginBottom: '0.25rem', color: 'var(--accent-black)' }}>{entity.name}</h1>
                        <p style={{ color: 'var(--text-secondary)', margin: 0, fontSize: '1.1rem' }}>{entity.sector}</p>
                    </div>
                </div>
                <div style={{ display: 'flex', gap: '1.5rem', alignItems: 'center' }}>
                    <span className={`badge ${entity.status === 'Compliant' ? 'badge-success' : entity.status === 'At Risk' ? 'badge-warning' : 'badge-danger'}`} style={{ fontSize: '1rem', padding: '10px 20px' }}>
                        {entity.status} (Score: {entity.score})
                    </span>
                    <button className="btn btn-primary" style={{ padding: '14px 28px', fontSize: '1rem' }}>Generate Proof Pack</button>
                </div>
            </header>

            <div style={{ display: 'flex', gap: '1rem', marginBottom: '2.5rem', overflowX: 'auto', paddingBottom: '0.5rem' }}>
                {tabs.map(tab => (
                    <button
                        key={tab}
                        onClick={() => setActiveTab(tab)}
                        style={{
                            background: activeTab === tab ? 'var(--accent-black)' : 'var(--glass-bg)',
                            border: '1px solid ' + (activeTab === tab ? 'transparent' : 'var(--glass-border)'),
                            color: activeTab === tab ? 'white' : 'var(--text-secondary)',
                            fontSize: '1rem',
                            fontWeight: 500,
                            cursor: 'pointer',
                            borderRadius: '100px',
                            padding: '12px 28px',
                            transition: 'all 0.2s',
                            whiteSpace: 'nowrap',
                            boxShadow: activeTab === tab ? '0 4px 12px rgba(0,0,0,0.15)' : 'none'
                        }}
                    >
                        {tab}
                    </button>
                ))}
            </div>

            <div className="glass-panel" style={{ padding: '3.5rem', minHeight: '500px' }}>
                <h2 style={{ fontSize: '2rem', marginBottom: '1.5rem', color: 'var(--accent-black)' }}>{activeTab}</h2>

                {activeTab === 'Obligation Mapper' && (
                    <div className="animate-slide-up">
                        <p style={{ fontSize: '1.1rem', color: 'var(--text-secondary)', marginBottom: '3rem', maxWidth: '800px' }}>
                            Translates {entity.sector} regulations into entity/site obligations automatically.
                        </p>
                        <div className="glass-panel-dark" style={{ padding: '2.5rem', marginBottom: '3rem' }}>
                            <h3 style={{ marginBottom: '1rem', color: '#94a3b8' }}>Active Obligations</h3>
                            <div style={{ fontSize: '1.4rem', lineHeight: 1.6, fontWeight: 500, color: 'var(--accent-white)' }}>{entity.obligations}</div>
                        </div>
                        <div>
                            <h3 style={{ marginBottom: '1rem', fontSize: '1.3rem', color: 'var(--accent-black)' }}>Group-to-Site Workflow Task</h3>
                            <div style={{ background: 'var(--bg-primary)', padding: '2rem', borderRadius: '16px', border: '1px solid var(--border-color)' }}>
                                <strong style={{ display: 'block', fontSize: '1rem', textTransform: 'uppercase', letterSpacing: '0.05em', color: 'var(--text-secondary)', marginBottom: '1rem' }}>AI Insight for {entity.name}:</strong>
                                <p style={{ fontStyle: 'italic', color: 'var(--accent-black)', fontSize: '1.3rem', margin: 0, lineHeight: 1.5 }}>"{entity.aiInsight}"</p>
                            </div>
                        </div>
                    </div>
                )}

                {activeTab === 'Evidence Vault' && (
                    <div className="animate-slide-up">
                        <p style={{ fontSize: '1.1rem', color: 'var(--text-secondary)', marginBottom: '3rem' }}>
                            Stores proofs, certificates, disclosures, test docs, declarations for {entity.name}.
                        </p>
                        <div style={{ padding: '2.5rem', background: 'rgba(239, 68, 68, 0.05)', borderRadius: '20px', border: '1px solid rgba(239, 68, 68, 0.2)' }}>
                            <div style={{ display: 'flex', alignItems: 'center', gap: '1rem', marginBottom: '1rem' }}>
                                <div style={{ width: '12px', height: '12px', borderRadius: '50%', background: 'var(--accent-danger)' }}></div>
                                <strong style={{ fontSize: '1.2rem', color: 'var(--accent-danger)', textTransform: 'uppercase', letterSpacing: '0.05em' }}>Evidence Status Warning</strong>
                            </div>
                            <p style={{ fontSize: '1.4rem', color: 'var(--accent-black)', fontWeight: 500 }}>{entity.evidenceStatus}</p>
                        </div>
                    </div>
                )}

                {activeTab === 'Supplier Outreach Engine' && (
                    <div className="animate-slide-up">
                        <p style={{ fontSize: '1.1rem', color: 'var(--text-secondary)', marginBottom: '3rem' }}>
                            Sends questionnaires and tracks missing evidence systematically across suppliers.
                        </p>
                        <div style={{ display: 'flex', gap: '1.5rem' }}>
                            <button className="btn btn-primary" style={{ padding: '16px 32px', fontSize: '1.1rem' }}>Automate Outreach</button>
                            <button className="btn btn-secondary" style={{ padding: '16px 32px', fontSize: '1.1rem' }}>Review Pending Requests</button>
                        </div>
                    </div>
                )}

                {activeTab === 'Audit Workspace' && (
                    <div className="animate-slide-up">
                        <p style={{ fontSize: '1.1rem', color: 'var(--text-secondary)', marginBottom: '3rem' }}>
                            Compiles regulator, customer, or internal audit response packs accurately.
                        </p>
                        <div style={{ height: '240px', display: 'flex', alignItems: 'center', justifyContent: 'center', background: 'var(--bg-primary)', borderRadius: '20px', border: '2px dashed var(--border-color)' }}>
                            <span style={{ color: 'var(--text-secondary)', fontWeight: 600, fontSize: '1.2rem', textTransform: 'uppercase', letterSpacing: '0.05em' }}>Immutable Audit Trail Active</span>
                        </div>
                    </div>
                )}
            </div>
        </div>
    );
};

export default EntityDetail;
