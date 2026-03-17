const ModuleView = ({ title }: { title: string }) => {
    return (
        <div className="content-area animate-slide-up">
            <header style={{ marginBottom: '2.5rem' }}>
                <h1 style={{ fontSize: '2.5rem', marginBottom: '0.5rem', color: 'var(--accent-black)' }}>{title}</h1>
                <p style={{ color: 'var(--text-secondary)', fontSize: '1.2rem', maxWidth: '800px' }}>
                    Group-wide {title.toLowerCase()} configurations and real-time operations.
                </p>
            </header>

            <div className="glass-panel" style={{ padding: '4rem', display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', minHeight: '500px', textAlign: 'center' }}>
                <div style={{ width: '80px', height: '80px', borderRadius: '50%', background: 'var(--bg-primary)', border: '1px solid var(--border-color)', display: 'flex', alignItems: 'center', justifyContent: 'center', marginBottom: '1.5rem', boxShadow: '0 8px 24px rgba(0,0,0,0.05)' }}>
                    <svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="var(--accent-black)" strokeWidth="2"><path d="M22 12h-4l-3 9L9 3l-3 9H2" /></svg>
                </div>
                <h2 style={{ fontSize: '1.8rem', color: 'var(--accent-black)', marginBottom: '1rem' }}>{title} Initialization</h2>
                <p style={{ color: 'var(--text-secondary)', maxWidth: '500px', fontSize: '1.1rem', lineHeight: 1.6 }}>
                    The {title} module is currently synchronizing federated group intelligence feeds for the Aditya Birla Group. Enterprise-scale risk and compliance data is being securely aggregated.
                </p>
                <button className="btn btn-primary" style={{ marginTop: '2.5rem', padding: '14px 28px', fontSize: '1.05rem' }}>Configure Module Access</button>
            </div>
        </div>
    );
};

export default ModuleView;
