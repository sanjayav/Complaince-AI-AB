import { useState } from 'react';

const Login = ({ onLogin }: { onLogin: (email: string) => void }) => {
    const [email, setEmail] = useState('');
    const [password, setPassword] = useState('');
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState('');

    const handleSubmit = (e: React.FormEvent) => {
        e.preventDefault();
        if (!email || !password) {
            setError('Please enter both email and password.');
            return;
        }
        setError('');
        setLoading(true);
        setTimeout(() => {
            setLoading(false);
            onLogin(email);
        }, 800);
    };

    return (
        <div style={{
            height: '100vh',
            width: '100vw',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            background: 'radial-gradient(circle at 20% 50%, #e0e6ed 0%, #f9fafb 40%, #dde3ec 100%)',
            position: 'relative',
            overflow: 'hidden',
        }}>
            {/* Decorative blobs */}
            <div style={{ position: 'absolute', top: '-120px', right: '-80px', width: '400px', height: '400px', borderRadius: '50%', background: 'rgba(15,23,42,0.04)', filter: 'blur(60px)' }} />
            <div style={{ position: 'absolute', bottom: '-100px', left: '-60px', width: '350px', height: '350px', borderRadius: '50%', background: 'rgba(16,185,129,0.06)', filter: 'blur(60px)' }} />

            <div className="login-card" style={{
                display: 'flex',
                width: '940px',
                maxWidth: '95vw',
                minHeight: '560px',
                borderRadius: '32px',
                overflow: 'hidden',
                background: 'rgba(255,255,255,0.55)',
                backdropFilter: 'blur(32px)',
                WebkitBackdropFilter: 'blur(32px)',
                border: '1px solid rgba(255,255,255,0.6)',
                boxShadow: '0 24px 48px -12px rgba(0,0,0,0.15)',
            }}>
                {/* Left branding panel */}
                <div className="login-branding" style={{
                    flex: 1,
                    background: 'linear-gradient(135deg, #0f172a 0%, #1e293b 100%)',
                    padding: '3.5rem',
                    display: 'flex',
                    flexDirection: 'column',
                    justifyContent: 'space-between',
                    color: 'white',
                    position: 'relative',
                    overflow: 'hidden',
                }}>
                    <div style={{ position: 'absolute', top: '-40px', right: '-40px', width: '200px', height: '200px', borderRadius: '50%', background: 'rgba(16,185,129,0.12)', filter: 'blur(40px)' }} />
                    <div style={{ position: 'absolute', bottom: '40px', left: '-20px', width: '160px', height: '160px', borderRadius: '50%', background: 'rgba(59,130,246,0.1)', filter: 'blur(40px)' }} />

                    <div style={{ position: 'relative', zIndex: 1 }}>
                        <div style={{ display: 'flex', alignItems: 'center', gap: '14px', marginBottom: '3rem' }}>
                            <div style={{ width: '48px', height: '48px', background: 'rgba(255,255,255,0.1)', borderRadius: '14px', display: 'flex', alignItems: 'center', justifyContent: 'center', border: '1px solid rgba(255,255,255,0.15)' }}>
                                <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="white" strokeWidth="2"><path d="M12 2L2 7l10 5 10-5-10-5zM2 17l10 5 10-5M2 12l10 5 10-5" /></svg>
                            </div>
                            <span style={{ fontSize: '1.6rem', fontWeight: 700, fontFamily: 'Outfit, sans-serif' }}>Aeiforo</span>
                        </div>

                        <h1 style={{ fontSize: '2.2rem', fontWeight: 700, lineHeight: 1.2, marginBottom: '1.25rem', fontFamily: 'Outfit, sans-serif', color: 'white' }}>
                            All your regulations.<br />
                            One AI Copilot.
                        </h1>
                        <p style={{ fontSize: '1.05rem', color: '#94a3b8', lineHeight: 1.7, maxWidth: '340px' }}>
                            Enterprise compliance intelligence for diversified industrial groups. Audit-ready answers in seconds.
                        </p>
                    </div>

                    <div style={{ position: 'relative', zIndex: 1 }}>
                        <div style={{ display: 'flex', gap: '2rem', marginBottom: '2rem' }}>
                            <div>
                                <div style={{ fontSize: '1.8rem', fontWeight: 700 }}>48</div>
                                <div style={{ fontSize: '0.8rem', color: '#64748b' }}>Frameworks</div>
                            </div>
                            <div>
                                <div style={{ fontSize: '1.8rem', fontWeight: 700 }}>24</div>
                                <div style={{ fontSize: '0.8rem', color: '#64748b' }}>Jurisdictions</div>
                            </div>
                            <div>
                                <div style={{ fontSize: '1.8rem', fontWeight: 700 }}>99.9%</div>
                                <div style={{ fontSize: '0.8rem', color: '#64748b' }}>Uptime</div>
                            </div>
                        </div>
                        <div style={{ display: 'flex', gap: '0.5rem' }}>
                            <span style={{ padding: '5px 12px', borderRadius: '100px', fontSize: '0.7rem', fontWeight: 600, background: 'rgba(16,185,129,0.15)', color: '#6ee7b7', border: '1px solid rgba(16,185,129,0.2)', textTransform: 'uppercase', letterSpacing: '0.05em' }}>SOC 2</span>
                            <span style={{ padding: '5px 12px', borderRadius: '100px', fontSize: '0.7rem', fontWeight: 600, background: 'rgba(59,130,246,0.15)', color: '#93c5fd', border: '1px solid rgba(59,130,246,0.2)', textTransform: 'uppercase', letterSpacing: '0.05em' }}>ISO 27001</span>
                            <span style={{ padding: '5px 12px', borderRadius: '100px', fontSize: '0.7rem', fontWeight: 600, background: 'rgba(255,255,255,0.08)', color: '#94a3b8', border: '1px solid rgba(255,255,255,0.1)', textTransform: 'uppercase', letterSpacing: '0.05em' }}>GDPR</span>
                        </div>
                    </div>
                </div>

                {/* Right login form */}
                <div className="login-form-panel" style={{
                    flex: 1,
                    padding: '3.5rem',
                    display: 'flex',
                    flexDirection: 'column',
                    justifyContent: 'center',
                }}>
                    <div style={{ maxWidth: '360px', width: '100%', margin: '0 auto' }}>
                        <h2 style={{ fontSize: '1.8rem', fontWeight: 700, marginBottom: '0.5rem', fontFamily: 'Outfit, sans-serif' }}>Welcome back</h2>
                        <p style={{ color: '#6b7280', fontSize: '1rem', marginBottom: '2.5rem' }}>Sign in to your compliance workspace</p>

                        <form onSubmit={handleSubmit}>
                            <div style={{ marginBottom: '1.25rem' }}>
                                <label style={{ display: 'block', fontSize: '0.85rem', fontWeight: 600, color: '#374151', marginBottom: '0.5rem' }}>Email address</label>
                                <input
                                    type="email"
                                    value={email}
                                    onChange={e => setEmail(e.target.value)}
                                    placeholder="you@company.com"
                                    style={{
                                        width: '100%',
                                        padding: '14px 16px',
                                        borderRadius: '14px',
                                        border: '1px solid rgba(0,0,0,0.08)',
                                        background: 'rgba(255,255,255,0.8)',
                                        fontSize: '1rem',
                                        outline: 'none',
                                        transition: 'all 0.2s',
                                        fontFamily: 'Inter, sans-serif',
                                    }}
                                    onFocus={e => { e.currentTarget.style.borderColor = '#0f172a'; e.currentTarget.style.boxShadow = '0 0 0 3px rgba(15,23,42,0.08)'; }}
                                    onBlur={e => { e.currentTarget.style.borderColor = 'rgba(0,0,0,0.08)'; e.currentTarget.style.boxShadow = 'none'; }}
                                />
                            </div>

                            <div style={{ marginBottom: '1.5rem' }}>
                                <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '0.5rem' }}>
                                    <label style={{ fontSize: '0.85rem', fontWeight: 600, color: '#374151' }}>Password</label>
                                    <span style={{ fontSize: '0.8rem', color: '#6b7280', cursor: 'pointer' }}>Forgot password?</span>
                                </div>
                                <input
                                    type="password"
                                    value={password}
                                    onChange={e => setPassword(e.target.value)}
                                    placeholder="Enter your password"
                                    style={{
                                        width: '100%',
                                        padding: '14px 16px',
                                        borderRadius: '14px',
                                        border: '1px solid rgba(0,0,0,0.08)',
                                        background: 'rgba(255,255,255,0.8)',
                                        fontSize: '1rem',
                                        outline: 'none',
                                        transition: 'all 0.2s',
                                        fontFamily: 'Inter, sans-serif',
                                    }}
                                    onFocus={e => { e.currentTarget.style.borderColor = '#0f172a'; e.currentTarget.style.boxShadow = '0 0 0 3px rgba(15,23,42,0.08)'; }}
                                    onBlur={e => { e.currentTarget.style.borderColor = 'rgba(0,0,0,0.08)'; e.currentTarget.style.boxShadow = 'none'; }}
                                />
                            </div>

                            {error && (
                                <div style={{ padding: '10px 14px', borderRadius: '10px', background: 'rgba(239,68,68,0.08)', color: '#dc2626', fontSize: '0.85rem', marginBottom: '1rem', border: '1px solid rgba(239,68,68,0.15)' }}>
                                    {error}
                                </div>
                            )}

                            <button
                                type="submit"
                                disabled={loading}
                                style={{
                                    width: '100%',
                                    padding: '14px',
                                    borderRadius: '14px',
                                    border: 'none',
                                    background: loading ? '#475569' : '#0f172a',
                                    color: 'white',
                                    fontSize: '1rem',
                                    fontWeight: 600,
                                    cursor: loading ? 'wait' : 'pointer',
                                    transition: 'all 0.2s',
                                    boxShadow: '0 4px 12px rgba(0,0,0,0.2)',
                                    fontFamily: 'Inter, sans-serif',
                                    display: 'flex',
                                    alignItems: 'center',
                                    justifyContent: 'center',
                                    gap: '8px',
                                }}
                            >
                                {loading && (
                                    <div style={{
                                        width: '18px', height: '18px', border: '2px solid rgba(255,255,255,0.3)',
                                        borderTopColor: 'white', borderRadius: '50%',
                                        animation: 'spin 0.6s linear infinite',
                                    }} />
                                )}
                                {loading ? 'Signing in...' : 'Sign in'}
                            </button>
                        </form>

                        <div style={{ display: 'flex', alignItems: 'center', gap: '1rem', margin: '1.75rem 0' }}>
                            <div style={{ flex: 1, height: '1px', background: 'rgba(0,0,0,0.08)' }} />
                            <span style={{ fontSize: '0.8rem', color: '#9ca3af' }}>or continue with</span>
                            <div style={{ flex: 1, height: '1px', background: 'rgba(0,0,0,0.08)' }} />
                        </div>

                        <div style={{ display: 'flex', gap: '0.75rem' }}>
                            <button onClick={() => { setEmail('demo@aeiforo.com'); setPassword('demo'); }}
                                style={{ flex: 1, padding: '12px', borderRadius: '14px', border: '1px solid rgba(0,0,0,0.08)', background: 'rgba(255,255,255,0.7)', cursor: 'pointer', fontSize: '0.9rem', fontWeight: 500, color: '#374151', fontFamily: 'Inter, sans-serif', transition: 'all 0.2s' }}>
                                SSO / Okta
                            </button>
                            <button onClick={() => { setEmail('demo@aeiforo.com'); setPassword('demo'); }}
                                style={{ flex: 1, padding: '12px', borderRadius: '14px', border: '1px solid rgba(0,0,0,0.08)', background: 'rgba(255,255,255,0.7)', cursor: 'pointer', fontSize: '0.9rem', fontWeight: 500, color: '#374151', fontFamily: 'Inter, sans-serif', transition: 'all 0.2s' }}>
                                Microsoft
                            </button>
                        </div>

                        <p style={{ textAlign: 'center', fontSize: '0.8rem', color: '#9ca3af', marginTop: '2rem' }}>
                            Protected by enterprise-grade encryption
                        </p>
                    </div>
                </div>
            </div>

            <style>{`
                @keyframes spin {
                    to { transform: rotate(360deg); }
                }
                @media (max-width: 768px) {
                    .login-card {
                        flex-direction: column !important;
                        width: 100% !important;
                        max-width: 100vw !important;
                        min-height: 100vh !important;
                        border-radius: 0 !important;
                        border: none !important;
                        box-shadow: none !important;
                    }
                    .login-branding {
                        padding: 2rem 1.5rem !important;
                        min-height: auto !important;
                    }
                    .login-branding h1 {
                        font-size: 1.5rem !important;
                    }
                    .login-form-panel {
                        padding: 2rem 1.5rem !important;
                    }
                }
            `}</style>
        </div>
    );
};

export default Login;
