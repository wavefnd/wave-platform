export type SourceCommit = {
  oid: string
  shortOid: string
  author: string
  authoredAt: string
  subject: string
}

export type SourceChangedFile = {
  status: string
  path: string
  oldPath: string
}

export type SourceCommitDetail = {
  commit: SourceCommit
  body: string
  parents: string[]
  files: SourceChangedFile[]
  patch: string
  patchTruncated: boolean
}

export type SourceRepository = {
  id: string
  owner: string
  name: string
  description: string
  defaultBranch: string
  headOid: string
  status: string
  headCommit?: SourceCommit
}

export type SourceEntry = {
  name: string
  path: string
  type: 'tree' | 'blob'
  oid: string
  size: number
  lastCommit?: SourceCommit
}

export type SourceLanguage = {
  name: string
  bytes: number
  files: number
  percent: number
  color: string
}

export type SourceBlob = {
  path: string
  oid: string
  size: number
  binary: boolean
  truncated: boolean
  content: string
  waveHighlight?: {
    engine: string
    abi: number
    tokens: Array<{
      kind: 'keyword' | 'type' | 'string' | 'comment' | 'number'
      start: number
      end: number
    }>
  }
}

export type SourceTree = {
  repository: SourceRepository
  ref: string
  path: string
  commit: SourceCommit
  entries: SourceEntry[]
  readme?: SourceBlob
  languages: SourceLanguage[]
}

export type SourceRef = { name: string; oid: string; updatedAt: string }
export type SourceRefs = { branches: SourceRef[]; tags: SourceRef[] }
