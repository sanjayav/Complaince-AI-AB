import { useState } from 'react';
import Login from './components/Login';
import Sidebar from './components/Sidebar';
import Dashboard from './components/Dashboard';
import BusinessUnits from './components/BusinessUnits';
import EntityDetail from './components/EntityDetail';
import ChatPanel from './components/ChatPanel';
import type { BusinessUnit } from './data';
import './index.css';

const alerts = [
    { id: '1', title: 'EU CBAM: New Reporting Format', impact: 'High', date: '2026-03-15', status: 'Action Required', description: 'Updated reporting requirements for carbon-intensive imports into the EU. Affects aluminum and cement exports.' },
    { id: '2', title: 'India BRSR: Scope Expansion', impact: 'Medium', date: '2026-03-10', status: 'In Review', description: 'SEBI proposes extending BRSR Core to top 2000 listed entities. Additional supply chain disclosures required.' },
    { id: '3', title: 'US SEC: Climate Disclosures Stayed', impact: 'Low', date: '2026-03-05', status: 'Monitoring', description: 'Eighth Circuit Court of Appeals issued an administrative stay on the SEC climate disclosure rules.' },
    { id: '4', title: 'EU CSDDD: Phased Rollout Starts', impact: 'High', date: '2026-02-28', status: 'Action Required', description: 'Corporate Sustainability Due Diligence Directive requires immediate supplier mapping for Group tier-1 companies.' },
];

const obligations = [
    { id: 'OBL-001', framework: 'GRI', requirement: 'GRI 305: Emissions 2016', mapping: 'Direct (Scope 1) GHG emissions', bu: 'Hindalco', status: 'Mapped', coverage: '100%' },
    { id: 'OBL-002', framework: 'BRSR', requirement: 'Principle 6', mapping: 'Energy and GHG Emissions', bu: 'Novelis', status: 'Mapped', coverage: '95%' },
    { id: 'OBL-003', framework: 'CSRD', requirement: 'Article 15', mapping: 'Adverse environmental impacts', bu: 'Aditya Birla Fashion', status: 'Partial', coverage: '60%' },
    { id: 'OBL-004', framework: 'ISSB', requirement: 'ESRS E1', mapping: 'Climate change mitigation', bu: 'UltraTech Cement', status: 'Unmapped', coverage: '0%' },
];

const documents = [
    { id: 'DOC-1029', name: 'Hindalco_Q3_Emissions_Report.pdf', type: 'Report', date: '2026-03-12', bu: 'Hindalco', status: 'Verified', aiConfidence: '98%' },
    { id: 'DOC-1030', name: 'Supplier_Code_of_Conduct_2026.docx', type: 'Policy', date: '2026-03-10', bu: 'Group Wide', status: 'Draft', aiConfidence: 'N/A' },
    { id: 'DOC-1031', name: 'UltraTech_Water_Usage_Log.xlsx', type: 'Data Log', date: '2026-03-08', bu: 'UltraTech Cement', status: 'Flagged', aiConfidence: '64%' },
    { id: 'DOC-1032', name: 'Novelis_Recycling_Certificates_Batch4.zip', type: 'Evidence', date: '2026-03-05', bu: 'Novelis', status: 'Verified', aiConfidence: '95%' },
];

const suppliers = [
    { id: 'SUP-0842', name: 'Global Logistics Corp', tier: 'Tier 1', risk: 'High', responseRate: '45%', status: 'At Risk', nextAction: 'Escalate' },
    { id: 'SUP-0193', name: 'Alumina Sources Ltd', tier: 'Tier 1', risk: 'Medium', responseRate: '100%', status: 'Compliant', nextAction: 'None' },
    { id: 'SUP-2041', name: 'TechPack Packaging', tier: 'Tier 2', risk: 'Low', responseRate: '85%', status: 'Pending Review', nextAction: 'Review Evidence' },
    { id: 'SUP-0522', name: 'Industrial ChemCo', tier: 'Tier 1', risk: 'High', responseRate: '12%', status: 'Non-Responsive', nextAction: 'Send Reminder' },
];

const audits = [
    { id: 'AUD-2026-Q1', title: 'Q1 Corporate Sustainability Audit', framework: 'BRSR', status: 'In Progress', progress: '75%', dueDate: '2026-03-31', assignee: 'Sarah Jenkins' },
    { id: 'AUD-2026-EU', title: 'CBAM Readiness Review', framework: 'EU CBAM', status: 'Planning', progress: '10%', dueDate: '2026-04-15', assignee: 'Michael Chang' },
    { id: 'AUD-2025-YE', title: 'FY25 Annual GHG Inventory', framework: 'GHG Protocol', status: 'Completed', progress: '100%', dueDate: '2026-01-31', assignee: 'Priya Sharma' },
];

const findings = [
    { id: 'FND-041', severity: 'Critical', description: 'Missing Scope 3 category 1 data for top 50 suppliers.', audit: 'Q1 Corporate Sustainability Audit' },
    { id: 'FND-042', severity: 'Medium', description: 'Inconsistent emission factors used across EU plants.', audit: 'CBAM Readiness Review' },
    { id: 'FND-043', severity: 'Low', description: 'Water usage meter calibration logs outdated for Plant B.', audit: 'Q1 Corporate Sustainability Audit' },
];

function renderContent(view: string) {
  switch (view) {
    case 'regulatory-radar':
      return (
        <div className="content-area animate-slide-up">
          <header style={{ marginBottom: '2.5rem' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
              <div>
                <h1 style={{ fontSize: '2.5rem', marginBottom: '0.5rem', color: 'var(--accent-black)' }}>Regulatory Radar</h1>
                <p style={{ color: 'var(--text-secondary)', fontSize: '1.2rem', maxWidth: '800px' }}>Real-time AI monitoring of global sustainability frameworks and compliance updates.</p>
              </div>
              <button className="btn btn-primary">+ Add Framework</button>
            </div>
          </header>
          <div className="module-stat-grid" style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: '1.5rem', marginBottom: '2.5rem' }}>
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
                <div key={alert.id} style={{ display: 'flex', flexWrap: 'wrap', gap: '1rem', paddingBottom: '1.5rem', borderBottom: '1px solid var(--border-color)', alignItems: 'flex-start' }}>
                  <div style={{ width: '10px', height: '10px', borderRadius: '50%', marginTop: '8px', flexShrink: 0, background: alert.impact === 'High' ? 'var(--accent-danger)' : alert.impact === 'Medium' ? 'var(--accent-warning)' : 'var(--accent-success)' }} />
                  <div style={{ flex: 1, minWidth: '200px' }}>
                    <div style={{ display: 'flex', flexWrap: 'wrap', justifyContent: 'space-between', marginBottom: '0.5rem', gap: '0.5rem' }}>
                      <h3 style={{ fontSize: '1.2rem', fontWeight: 600 }}>{alert.title}</h3>
                      <span style={{ fontSize: '0.9rem', color: 'var(--text-secondary)' }}>{alert.date}</span>
                    </div>
                    <p style={{ color: 'var(--text-secondary)', marginBottom: '1rem', lineHeight: 1.5 }}>{alert.description}</p>
                    <div style={{ display: 'flex', gap: '1rem', alignItems: 'center', flexWrap: 'wrap' }}>
                      <span className="badge" style={{ background: alert.status === 'Action Required' ? 'rgba(239,68,68,0.1)' : 'rgba(0,0,0,0.05)', color: alert.status === 'Action Required' ? 'var(--accent-danger)' : 'var(--accent-black)' }}>{alert.status}</span>
                      <span style={{ fontSize: '0.9rem', color: 'var(--text-secondary)', fontWeight: 500 }}>Impact: {alert.impact}</span>
                    </div>
                  </div>
                  <button className="btn btn-secondary" style={{ padding: '8px 16px', fontSize: '0.9rem', flexShrink: 0 }}>View Gap Analysis</button>
                </div>
              ))}
            </div>
          </div>
        </div>
      );
    case 'obligation-mapper':
      return (
        <div className="content-area animate-slide-up">
          <header style={{ marginBottom: '2.5rem' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
              <div>
                <h1 style={{ fontSize: '2.5rem', marginBottom: '0.5rem', color: 'var(--accent-black)' }}>Obligation Mapper</h1>
                <p style={{ color: 'var(--text-secondary)', fontSize: '1.2rem', maxWidth: '800px' }}>Automatically map regulatory requirements to internal business unit controls and data points.</p>
              </div>
              <button className="btn btn-primary">Run AI Mapping</button>
            </div>
          </header>
          <div className="module-stat-grid" style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: '1.5rem', marginBottom: '2.5rem' }}>
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
            <div className="table-scroll">
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
                {obligations.map((obl) => (
                  <tr key={obl.id} style={{ borderBottom: '1px solid var(--border-color)' }}>
                    <td style={{ padding: '1rem', fontWeight: 600 }}>{obl.id}</td>
                    <td style={{ padding: '1rem' }}><span style={{ background: 'var(--bg-primary)', padding: '4px 8px', borderRadius: '4px', fontSize: '0.85rem' }}>{obl.framework}</span></td>
                    <td style={{ padding: '1rem' }}>{obl.requirement}</td>
                    <td style={{ padding: '1rem' }}>{obl.mapping}</td>
                    <td style={{ padding: '1rem', color: 'var(--text-secondary)' }}>{obl.bu}</td>
                    <td style={{ padding: '1rem' }}>
                      <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                        <div style={{ width: '60px', height: '6px', background: 'rgba(0,0,0,0.1)', borderRadius: '3px', overflow: 'hidden' }}><div style={{ width: obl.coverage, height: '100%', background: obl.coverage === '100%' ? 'var(--accent-success)' : obl.coverage === '0%' ? 'var(--accent-danger)' : 'var(--accent-warning)' }} /></div>
                        <span style={{ fontSize: '0.85rem' }}>{obl.coverage}</span>
                      </div>
                    </td>
                    <td style={{ padding: '1rem' }}><span className="badge" style={{ background: obl.status === 'Mapped' ? 'rgba(34,197,94,0.1)' : obl.status === 'Partial' ? 'rgba(234,179,8,0.1)' : 'rgba(239,68,68,0.1)', color: obl.status === 'Mapped' ? 'var(--accent-success)' : obl.status === 'Partial' ? 'var(--accent-warning)' : 'var(--accent-danger)' }}>{obl.status}</span></td>
                  </tr>
                ))}
              </tbody>
            </table>
            </div>
          </div>
        </div>
      );
    case 'evidence-vault':
      return (
        <div className="content-area animate-slide-up">
          <header style={{ marginBottom: '2.5rem' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
              <div>
                <h1 style={{ fontSize: '2.5rem', marginBottom: '0.5rem', color: 'var(--accent-black)' }}>Evidence Vault</h1>
                <p style={{ color: 'var(--text-secondary)', fontSize: '1.2rem', maxWidth: '800px' }}>Centralized immutable storage for compliance documentation and automatically extracted metrics.</p>
              </div>
              <div style={{ display: 'flex', gap: '1rem' }}>
                <button className="btn btn-secondary">Bulk Upload</button>
                <button className="btn btn-primary">Upload Document</button>
              </div>
            </div>
          </header>
          <div className="module-stat-grid" style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: '1.5rem', marginBottom: '2.5rem' }}>
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
            <h2 style={{ fontSize: '1.5rem', marginBottom: '1.5rem' }}>Recent Ingestions</h2>
            <div className="table-scroll">
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
                </tr>
              </thead>
              <tbody>
                {documents.map((doc) => (
                  <tr key={doc.id} style={{ borderBottom: '1px solid var(--border-color)' }}>
                    <td style={{ padding: '1rem', fontSize: '0.9rem', color: 'var(--text-secondary)' }}>{doc.id}</td>
                    <td style={{ padding: '1rem', fontWeight: 500, color: 'var(--accent-black)' }}>{doc.name}</td>
                    <td style={{ padding: '1rem' }}>{doc.type}</td>
                    <td style={{ padding: '1rem' }}>{doc.bu}</td>
                    <td style={{ padding: '1rem' }}>{doc.date}</td>
                    <td style={{ padding: '1rem' }}><span style={{ color: parseInt(doc.aiConfidence) >= 90 ? 'var(--accent-success)' : parseInt(doc.aiConfidence) >= 50 ? 'var(--accent-warning)' : 'var(--text-secondary)' }}>{doc.aiConfidence}</span></td>
                    <td style={{ padding: '1rem' }}><span className="badge" style={{ background: doc.status === 'Verified' ? 'rgba(34,197,94,0.1)' : doc.status === 'Flagged' ? 'rgba(239,68,68,0.1)' : 'rgba(0,0,0,0.05)', color: doc.status === 'Verified' ? 'var(--accent-success)' : doc.status === 'Flagged' ? 'var(--accent-danger)' : 'var(--accent-black)' }}>{doc.status}</span></td>
                  </tr>
                ))}
              </tbody>
            </table>
            </div>
          </div>
        </div>
      );
    case 'supplier-engine':
      return (
        <div className="content-area animate-slide-up">
          <header style={{ marginBottom: '2.5rem' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
              <div>
                <h1 style={{ fontSize: '2.5rem', marginBottom: '0.5rem', color: 'var(--accent-black)' }}>Supplier Engine</h1>
                <p style={{ color: 'var(--text-secondary)', fontSize: '1.2rem', maxWidth: '800px' }}>Automated outreach, data collection, and risk scoring across the group supply chain.</p>
              </div>
              <button className="btn btn-secondary">Launch Campaign</button>
            </div>
          </header>
          <div className="module-stat-grid" style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: '1.5rem', marginBottom: '2.5rem' }}>
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
            <h2 style={{ fontSize: '1.5rem', marginBottom: '1.5rem' }}>Supplier Watchlist</h2>
            <div className="table-scroll">
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
                {suppliers.map((sup) => (
                  <tr key={sup.id} style={{ borderBottom: '1px solid var(--border-color)' }}>
                    <td style={{ padding: '1rem', fontWeight: 600 }}>{sup.name}</td>
                    <td style={{ padding: '1rem', color: 'var(--text-secondary)' }}>{sup.tier}</td>
                    <td style={{ padding: '1rem' }}><span style={{ color: sup.risk === 'High' ? 'var(--accent-danger)' : sup.risk === 'Medium' ? 'var(--accent-warning)' : 'var(--accent-success)', fontWeight: 600 }}>{sup.risk}</span></td>
                    <td style={{ padding: '1rem' }}>
                      <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                        <div style={{ width: '60px', height: '6px', background: 'rgba(0,0,0,0.1)', borderRadius: '3px', overflow: 'hidden' }}><div style={{ width: sup.responseRate, height: '100%', background: 'var(--accent-black)' }} /></div>
                        <span style={{ fontSize: '0.85rem' }}>{sup.responseRate}</span>
                      </div>
                    </td>
                    <td style={{ padding: '1rem' }}><span className="badge" style={{ background: sup.status === 'Compliant' ? 'rgba(34,197,94,0.1)' : sup.status === 'At Risk' || sup.status === 'Non-Responsive' ? 'rgba(239,68,68,0.1)' : 'rgba(234,179,8,0.1)', color: sup.status === 'Compliant' ? 'var(--accent-success)' : sup.status === 'At Risk' || sup.status === 'Non-Responsive' ? 'var(--accent-danger)' : 'var(--accent-warning)' }}>{sup.status}</span></td>
                    <td style={{ padding: '1rem' }}><button style={{ background: 'none', border: '1px solid var(--border-color)', padding: '4px 12px', borderRadius: '6px', fontSize: '0.85rem', cursor: 'pointer' }}>{sup.nextAction}</button></td>
                  </tr>
                ))}
              </tbody>
            </table>
            </div>
          </div>
        </div>
      );
    case 'audit-workspace':
      return (
        <div className="content-area animate-slide-up">
          <header style={{ marginBottom: '2.5rem' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
              <div>
                <h1 style={{ fontSize: '2.5rem', marginBottom: '0.5rem', color: 'var(--accent-black)' }}>Audit Workspace</h1>
                <p style={{ color: 'var(--text-secondary)', fontSize: '1.2rem', maxWidth: '800px' }}>Generate audit-ready evidence packs and track internal compliance readiness.</p>
              </div>
              <button className="btn btn-primary">+ New Audit Plan</button>
            </div>
          </header>
          <div className="module-stat-grid" style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: '1.5rem', marginBottom: '2.5rem' }}>
            {audits.map((audit) => (
              <div key={audit.id} className="glass-panel" style={{ padding: '1.5rem' }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: '1rem' }}>
                  <h3 style={{ fontSize: '1.1rem', fontWeight: 600 }}>{audit.title}</h3>
                  <span className="badge" style={{ background: audit.status === 'Completed' ? 'rgba(34,197,94,0.1)' : audit.status === 'In Progress' ? 'rgba(59,130,246,0.1)' : 'rgba(0,0,0,0.05)', color: audit.status === 'Completed' ? 'var(--accent-success)' : audit.status === 'In Progress' ? '#3b82f6' : 'var(--text-secondary)', flexShrink: 0 }}>{audit.status}</span>
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
                  <div style={{ width: '100%', height: '6px', background: 'rgba(0,0,0,0.1)', borderRadius: '3px', overflow: 'hidden' }}><div style={{ width: audit.progress, height: '100%', background: audit.progress === '100%' ? 'var(--accent-success)' : 'var(--accent-black)' }} /></div>
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
                <div key={finding.id} style={{ display: 'flex', flexWrap: 'wrap', gap: '1rem', padding: '1rem', border: '1px solid var(--border-color)', borderRadius: '12px', alignItems: 'center' }}>
                  <div style={{ width: '4px', height: '40px', borderRadius: '2px', flexShrink: 0, background: finding.severity === 'Critical' ? 'var(--accent-danger)' : finding.severity === 'Medium' ? 'var(--accent-warning)' : 'var(--accent-success)' }} />
                  <div style={{ flex: 1, minWidth: '200px' }}>
                    <p style={{ fontWeight: 500, marginBottom: '0.25rem' }}>{finding.description}</p>
                    <div style={{ display: 'flex', flexWrap: 'wrap', gap: '0.5rem', fontSize: '0.85rem', color: 'var(--text-secondary)' }}>
                      <span>{finding.id}</span>
                      <span>•</span>
                      <span>{finding.audit}</span>
                    </div>
                  </div>
                  <button className="btn btn-secondary" style={{ padding: '6px 12px', fontSize: '0.85rem', flexShrink: 0 }}>Resolve</button>
                </div>
              ))}
            </div>
          </div>
        </div>
      );
    default:
      return null;
  }
}

function App() {
  const [user, setUser] = useState<string | null>(null);
  const [currentView, setCurrentView] = useState<string>('dashboard');
  const [selectedEntity, setSelectedEntity] = useState<BusinessUnit | null>(null);
  const [chatOpen, setChatOpen] = useState(false);
  const [sidebarOpen, setSidebarOpen] = useState(false);

  const handleSelectEntity = (entity: BusinessUnit) => {
    setSelectedEntity(entity);
    setCurrentView('detail');
  };

  const handleBack = () => {
    setSelectedEntity(null);
    setCurrentView('entities');
  };

  if (!user) {
    return <Login onLogin={(email) => setUser(email)} />;
  }

  const moduleContent = renderContent(currentView);

  return (
    <div className="app-wrapper">
      {/* Mobile top header */}
      <div className="mobile-header">
        <div className="mobile-header-logo">
          <div style={{ width: '34px', height: '34px', background: 'var(--accent-black)', borderRadius: '10px', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="white" strokeWidth="2"><path d="M12 2L2 7l10 5 10-5-10-5zM2 17l10 5 10-5M2 12l10 5 10-5" /></svg>
          </div>
          <span style={{ fontWeight: 700, fontSize: '1.15rem', fontFamily: 'Outfit, sans-serif' }}>Aeiforo</span>
        </div>
        <button className="hamburger-btn" onClick={() => setSidebarOpen(true)} aria-label="Open menu">
          <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><line x1="3" y1="6" x2="21" y2="6" /><line x1="3" y1="12" x2="21" y2="12" /><line x1="3" y1="18" x2="21" y2="18" /></svg>
        </button>
      </div>

      {/* Mobile sidebar drawer */}
      <div className={`sidebar-mobile-wrap ${sidebarOpen ? 'open' : ''}`}>
        <div className="sidebar-overlay" onClick={() => setSidebarOpen(false)} />
        <div className="sidebar-drawer">
          <Sidebar currentView={currentView} onViewChange={setCurrentView} userEmail={user ?? undefined} onLogout={() => setUser(null)} onNavigate={() => setSidebarOpen(false)} />
        </div>
      </div>

      <div className="main-container">
        {/* Desktop sidebar */}
        <div className="sidebar-desktop">
          <Sidebar currentView={currentView} onViewChange={setCurrentView} userEmail={user ?? undefined} onLogout={() => setUser(null)} />
        </div>

        {currentView === 'dashboard' && <Dashboard />}
        {currentView === 'entities' && <BusinessUnits onSelect={handleSelectEntity} />}
        {currentView === 'detail' && selectedEntity && <EntityDetail entity={selectedEntity} onBack={handleBack} />}
        {moduleContent}
      </div>

      {/* Chat FAB */}
      {!chatOpen && (
        <button
          className="chat-fab"
          onClick={() => setChatOpen(true)}
          style={{
            position: 'fixed', bottom: '24px', right: '24px', width: '60px', height: '60px',
            borderRadius: '50%', border: 'none', background: '#0f172a', color: 'white',
            cursor: 'pointer', boxShadow: '0 8px 24px rgba(0,0,0,0.3)', display: 'flex',
            alignItems: 'center', justifyContent: 'center', zIndex: 999,
            transition: 'transform 0.2s, box-shadow 0.2s',
          }}
          onMouseEnter={e => { e.currentTarget.style.transform = 'scale(1.08)'; e.currentTarget.style.boxShadow = '0 12px 32px rgba(0,0,0,0.4)'; }}
          onMouseLeave={e => { e.currentTarget.style.transform = 'scale(1)'; e.currentTarget.style.boxShadow = '0 8px 24px rgba(0,0,0,0.3)'; }}
        >
          <svg width="26" height="26" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z" /></svg>
        </button>
      )}

      <ChatPanel isOpen={chatOpen} onClose={() => setChatOpen(false)} />
    </div>
  );
}

export default App;
