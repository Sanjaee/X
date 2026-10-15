export const feedKeys = {
  all: ["feed"] as const,
  list: (type: string) => [...feedKeys.all, "list", type] as const,
  detail: (id: string) => [...feedKeys.all, "detail", id] as const,
};

export const searchKeys = {
  all: ["search"] as const,
  posts: (query: string, tab: string, filters?: Record<string, unknown>) =>
    [...searchKeys.all, "posts", query, tab, filters ?? {}] as const,
  users: (query: string) => [...searchKeys.all, "users", query] as const,
  suggestions: (query: string) => [...searchKeys.all, "suggestions", query] as const,
  recent: () => [...searchKeys.all, "recent"] as const,
};
