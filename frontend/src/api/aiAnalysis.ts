import axios from 'axios'

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080'

export interface AnomalyDetectionRule {
  id: string
  userId: string
  clusterId?: string
  dataSourceId?: string
  name: string
  description?: string
  metricQuery: string
  algorithm: 'stl' | 'isolation_forest' | 'lstm' | 'baseline'
  sensitivity: number
  windowSize: number
  minValue?: number
  maxValue?: number
  enabled: boolean
  evalInterval: number
  lastEvalAt?: string
  alertThreshold: number
  alertOnRecovery: boolean
  notificationChannels?: string
  totalEvaluations: number
  anomaliesDetected: number
  lastAnomalyAt?: string
  createdAt: string
  updatedAt: string
  cluster?: {
    id: string
    name: string
  }
  dataSource?: {
    id: string
    name: string
  }
}

export interface CreateAnomalyDetectionRuleRequest {
  clusterId?: string
  dataSourceId?: string
  name: string
  description?: string
  metricQuery: string
  algorithm: 'stl' | 'isolation_forest' | 'lstm' | 'baseline'
  sensitivity?: number
  windowSize?: number
  minValue?: number
  maxValue?: number
  enabled?: boolean
  evalInterval?: number
  alertThreshold?: number
  alertOnRecovery?: boolean
  notificationChannels?: string
}

export interface UpdateAnomalyDetectionRuleRequest {
  name?: string
  description?: string
  metricQuery?: string
  algorithm?: 'stl' | 'isolation_forest' | 'lstm' | 'baseline'
  sensitivity?: number
  windowSize?: number
  minValue?: number
  maxValue?: number
  enabled?: boolean
  evalInterval?: number
  alertThreshold?: number
  alertOnRecovery?: boolean
  notificationChannels?: string
}

export interface ExecuteAnomalyDetectionRequest {
  ruleId: string
  startTime?: string
  endTime?: string
}

export interface ExecuteAnomalyDetectionResponse {
  anomalies: AnomalyEvent[]
  anomalyCount: number
  evaluatedAt: string
  duration: number
}

export interface AnomalyEvent {
  id: string
  ruleId: string
  userId: string
  clusterId?: string
  severity: 'critical' | 'warning' | 'info'
  metricName: string
  currentValue: number
  expectedValue: number
  deviation: number
  confidence: number
  timeRange?: string
  labels?: string
  description: string
  suggestions?: string
  status: 'active' | 'acknowledged' | 'resolved' | 'false_positive'
  acknowledgedAt?: string
  acknowledgedBy?: string
  resolvedAt?: string
  resolvedBy?: string
  createdAt: string
  updatedAt: string
  rule?: AnomalyDetectionRule
  cluster?: {
    id: string
    name: string
  }
}

export interface LLMConversation {
  id: string
  userId: string
  clusterId?: string
  title?: string
  model: string
  temperature: number
  maxTokens: number
  systemPrompt?: string
  relatedAlertIds?: string
  relatedAnomalyIds?: string
  createdAt: string
  updatedAt: string
  cluster?: {
    id: string
    name: string
  }
  messages?: LLMMessage[]
}

export interface LLMMessage {
  id: string
  conversationId: string
  role: 'user' | 'assistant' | 'system'
  content: string
  tokensUsed?: number
  relatedQuery?: string
  relatedData?: string
  createdAt: string
}

export interface CreateLLMConversationRequest {
  clusterId?: string
  title?: string
  model: string
  temperature?: number
  maxTokens?: number
  systemPrompt?: string
}

export interface SendLLMMessageRequest {
  content: string
}

export interface SendLLMMessageResponse {
  messageId: string
  content: string
  tokensUsed: number
  relatedQuery?: string
  relatedData?: string
}

const aiAnalysisApi = {
  // ============== Anomaly Detection Rules ==============

  // List all anomaly detection rules
  getAnomalyRules: async (params?: {
    clusterId?: string
    dataSourceId?: string
    algorithm?: string
    enabled?: string
    page?: number
    pageSize?: number
  }) => {
    const response = await axios.get<{
      data: AnomalyDetectionRule[]
      total: number
      page: number
      pageSize: number
      totalPages: number
    }>(`${API_BASE_URL}/api/v1/ai/anomaly-rules`, {
      headers: {
        Authorization: `Bearer ${localStorage.getItem('token')}`,
      },
      params,
    })
    return response.data
  },

  // Get a specific anomaly detection rule
  getAnomalyRule: async (id: string) => {
    const response = await axios.get<AnomalyDetectionRule>(
      `${API_BASE_URL}/api/v1/ai/anomaly-rules/${id}`,
      {
        headers: {
          Authorization: `Bearer ${localStorage.getItem('token')}`,
        },
      }
    )
    return response.data
  },

  // Create a new anomaly detection rule
  createAnomalyRule: async (data: CreateAnomalyDetectionRuleRequest) => {
    const response = await axios.post<AnomalyDetectionRule>(
      `${API_BASE_URL}/api/v1/ai/anomaly-rules`,
      data,
      {
        headers: {
          Authorization: `Bearer ${localStorage.getItem('token')}`,
        },
      }
    )
    return response.data
  },

  // Update an anomaly detection rule
  updateAnomalyRule: async (id: string, data: UpdateAnomalyDetectionRuleRequest) => {
    const response = await axios.put<AnomalyDetectionRule>(
      `${API_BASE_URL}/api/v1/ai/anomaly-rules/${id}`,
      data,
      {
        headers: {
          Authorization: `Bearer ${localStorage.getItem('token')}`,
        },
      }
    )
    return response.data
  },

  // Delete an anomaly detection rule
  deleteAnomalyRule: async (id: string) => {
    const response = await axios.delete<{ message: string }>(
      `${API_BASE_URL}/api/v1/ai/anomaly-rules/${id}`,
      {
        headers: {
          Authorization: `Bearer ${localStorage.getItem('token')}`,
        },
      }
    )
    return response.data
  },

  // Execute anomaly detection
  executeAnomalyDetection: async (data: ExecuteAnomalyDetectionRequest) => {
    const response = await axios.post<ExecuteAnomalyDetectionResponse>(
      `${API_BASE_URL}/api/v1/ai/anomaly-rules/execute`,
      data,
      {
        headers: {
          Authorization: `Bearer ${localStorage.getItem('token')}`,
        },
      }
    )
    return response.data
  },

  // ============== Anomaly Events ==============

  // List all anomaly events
  getAnomalyEvents: async (params?: {
    clusterId?: string
    ruleId?: string
    severity?: string
    status?: string
    page?: number
    pageSize?: number
  }) => {
    const response = await axios.get<{
      data: AnomalyEvent[]
      total: number
      page: number
      pageSize: number
      totalPages: number
    }>(`${API_BASE_URL}/api/v1/ai/anomaly-events`, {
      headers: {
        Authorization: `Bearer ${localStorage.getItem('token')}`,
      },
      params,
    })
    return response.data
  },

  // ============== LLM Conversations ==============

  // List all LLM conversations
  getLLMConversations: async (params?: {
    page?: number
    pageSize?: number
  }) => {
    const response = await axios.get<{
      data: LLMConversation[]
      total: number
      page: number
      pageSize: number
      totalPages: number
    }>(`${API_BASE_URL}/api/v1/ai/llm/conversations`, {
      headers: {
        Authorization: `Bearer ${localStorage.getItem('token')}`,
      },
      params,
    })
    return response.data
  },

  // Get a specific LLM conversation
  getLLMConversation: async (id: string) => {
    const response = await axios.get<LLMConversation>(
      `${API_BASE_URL}/api/v1/ai/llm/conversations/${id}`,
      {
        headers: {
          Authorization: `Bearer ${localStorage.getItem('token')}`,
        },
      }
    )
    return response.data
  },

  // Create a new LLM conversation
  createLLMConversation: async (data: CreateLLMConversationRequest) => {
    const response = await axios.post<LLMConversation>(
      `${API_BASE_URL}/api/v1/ai/llm/conversations`,
      data,
      {
        headers: {
          Authorization: `Bearer ${localStorage.getItem('token')}`,
        },
      }
    )
    return response.data
  },

  // Delete an LLM conversation
  deleteLLMConversation: async (id: string) => {
    const response = await axios.delete<{ message: string }>(
      `${API_BASE_URL}/api/v1/ai/llm/conversations/${id}`,
      {
        headers: {
          Authorization: `Bearer ${localStorage.getItem('token')}`,
        },
      }
    )
    return response.data
  },

  // Send a message in an LLM conversation
  sendLLMMessage: async (conversationId: string, data: SendLLMMessageRequest) => {
    const response = await axios.post<SendLLMMessageResponse>(
      `${API_BASE_URL}/api/v1/ai/llm/conversations/${conversationId}/messages`,
      data,
      {
        headers: {
          Authorization: `Bearer ${localStorage.getItem('token')}`,
        },
      }
    )
    return response.data
  },
}

export default aiAnalysisApi
