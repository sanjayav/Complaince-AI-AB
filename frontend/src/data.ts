export interface BusinessUnit {
  id: string;
  name: string;
  sector: string;
  description: string;
  status: 'Compliant' | 'Action Required' | 'At Risk';
  score: number;
  obligations: string;
  evidenceStatus: string;
  aiInsight: string;
  modules: string[];
}

export const businessUnits: BusinessUnit[] = [
  {
    id: 'hindalco',
    name: 'Hindalco',
    sector: 'Metals / Aluminium',
    description: 'Exporting aluminium products to Europe. Focus on CBAM evidence and emissions data management across plant shipments.',
    status: 'Action Required',
    score: 82,
    obligations: 'CBAM Declarations, EU ETS alignment.',
    evidenceStatus: 'Missing Scope 3 emissions data for Q1 EU shipments.',
    aiInsight: 'We export aluminium products to Europe. What CBAM evidence and emissions data must be maintained by plant and shipment?',
    modules: ['Regulatory Radar', 'Gap & Exposure Engine']
  },
  {
    id: 'ultratech',
    name: 'UltraTech',
    sector: 'Cement',
    description: 'Decarbonisation tracking and sustainability disclosure management for climate reporting.',
    status: 'Compliant',
    score: 95,
    obligations: 'CSRD, Taskforce on Climate-Related Financial Disclosures (TCFD).',
    evidenceStatus: 'All plant-level emissions workflows uploaded & verified.',
    aiInsight: 'Which decarbonisation, disclosure, and plant evidence workflows are needed to support climate reporting and customer questionnaires?',
    modules: ['Obligation Mapper', 'Audit Workspace']
  },
  {
    id: 'birlacarbon',
    name: 'Birla Carbon',
    sector: 'Chemicals / Materials',
    description: 'Substantiating sustainability claims and product circularity inputs across regions.',
    status: 'At Risk',
    score: 68,
    obligations: 'Green Claims Directive, REACH.',
    evidenceStatus: 'Pending customer-grade documentation for circular inputs from 3 APAC suppliers.',
    aiInsight: 'How do we substantiate sustainability claims, product circularity inputs, and customer-grade documentation across regions?',
    modules: ['Supplier Outreach Engine', 'Evidence Vault']
  },
  {
    id: 'textiles',
    name: 'Aditya Birla Textiles',
    sector: 'Fibre / Fashion',
    description: 'Responding to customer due diligence and restricted substance data requests.',
    status: 'Action Required',
    score: 75,
    obligations: 'EU CSDDD, ZDHC (Zero Discharge of Hazardous Chemicals).',
    evidenceStatus: 'Awaiting source-linked evidence from 12 tier-2 suppliers.',
    aiInsight: 'How do we respond to customer due diligence, restricted substance, traceability, and sustainability data requests with source-linked evidence?',
    modules: ['Supplier Outreach Engine', 'Gap & Exposure Engine']
  }
];
