export interface RetryMap {
  [childSessionId: string]: string; // childSessionId -> parentSessionId
}

const RETRY_MAP_KEY = "quizgen_retry_map_v1";

export function getRetryMap(): RetryMap {
  if (typeof window === "undefined") return {};
  try {
    const data = localStorage.getItem(RETRY_MAP_KEY);
    return data ? JSON.parse(data) : {};
  } catch (e) {
    return {};
  }
}

export function saveRetryMapping(parentSessionId: string, childSessionId: string) {
  if (typeof window === "undefined") return;
  try {
    const current = getRetryMap();
    current[childSessionId] = parentSessionId;
    localStorage.setItem(RETRY_MAP_KEY, JSON.stringify(current));
  } catch (e) {
    console.error("Failed to save retry mapping", e);
  }
}

export function getParentSessionId(childSessionId: string): string | null {
  const map = getRetryMap();
  return map[childSessionId] || null;
}

export function getChildSessionIds(parentSessionId: string): string[] {
  const map = getRetryMap();
  const children: string[] = [];
  for (const [childId, parentId] of Object.entries(map)) {
    if (parentId === parentSessionId) {
      children.push(childId);
    }
  }
  return children;
}
