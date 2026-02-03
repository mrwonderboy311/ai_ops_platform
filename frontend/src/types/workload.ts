// Workload types for Kubernetes resources

export interface K8sNamespace {
  name: string
}

export interface K8sDeployment {
  name: string
  namespace: string
  replicas: number
  availableReplicas: number
  readyReplicas: number
  updatedReplicas: number
  labels: Record<string, string>
  selector: Record<string, string>
  image: string
}

export interface K8sPod {
  name: string
  namespace: string
  status: string
  phase: string
  nodeName: string
  hostIp: string
  podIp: string
  ready: boolean
  restartCount: number
  labels: Record<string, string>
  ownerType: string
  ownerName: string
  containers: K8sContainer[]
  createdAt: string
}

export interface K8sContainer {
  name: string
  image: string
  ready: boolean
  restartCount: number
}

export interface K8sService {
  name: string
  namespace: string
  type: string
  clusterIp: string
  externalIps: string[]
  ports: K8sServicePort[]
  selector: Record<string, string>
  labels: Record<string, string>
  createdAt: string
}

export interface K8sServicePort {
  name: string
  port: number
  targetPort: number
  protocol: string
  nodePort?: number
}

export interface PodLogsResponse {
  data: {
    podName: string
    logs: string
  }
}

export interface NamespacesResponse {
  data: string[]
}

export interface DeploymentsResponse {
  data: K8sDeployment[]
}

export interface PodsResponse {
  data: K8sPod[]
}

export interface ServicesResponse {
  data: K8sService[]
}
