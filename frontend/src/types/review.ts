export interface ReviewMetric {
  key?: string;
  label: string;
  value: string;
  hint: string;
}

export interface ReviewInsight {
  title: string;
  evidence: string;
  suggestion: string;
}

export interface ReviewSummary {
  metrics: ReviewMetric[];
  insights: ReviewInsight[];
  sampleWarning?: string;
}
