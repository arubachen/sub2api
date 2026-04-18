interface APIErrorLike {
  message?: string
  response?: {
    data?: {
      detail?: string
      message?: string
      reason?: string
    }
  }
}

function extractErrorMessage(error: unknown): string {
  const err = (error || {}) as APIErrorLike
  return err.response?.data?.detail || err.response?.data?.message || err.message || ''
}

function extractErrorReason(error: unknown): string {
  const err = (error || {}) as APIErrorLike
  return err.response?.data?.reason || ''
}

export function buildAuthErrorMessage(
  error: unknown,
  options: {
    fallback: string
    reasonMap?: Record<string, string>
  }
): string {
  const { fallback, reasonMap } = options
  const reason = extractErrorReason(error)
  if (reason && reasonMap?.[reason]) {
    return reasonMap[reason]
  }
  const message = extractErrorMessage(error)
  return message || fallback
}
