// API client for Memento
// Based on swagger.yaml API definitions

declare const process: {
  env: {
    NEXT_PUBLIC_API_URL?: string;
  }
};

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || '/api/v1';

// Common headers with auth token
const getAuthHeaders = (): HeadersInit => {
  const token = typeof window !== 'undefined' ? localStorage.getItem("mementoToken") : null;
  return {
    'Content-Type': 'application/json',
    ...(token ? { 'Authorization': `Bearer ${token}` } : {}),
  };
};

// Error handler
const handleResponse = async (response: Response) => {
  console.log(`API Client: Handling response from ${response.url}, status: ${response.status}`);
  
  if (!response.ok) {
    try {
      const data = await response.json();
      console.error(`API Error: ${response.status} ${response.statusText}`, data);
      // Handle error responses
      const error = new Error(data.message || 'API request failed');
      throw Object.assign(error, { status: response.status, data });
    } catch (jsonError) {
      console.error(`API Error: ${response.status} ${response.statusText} (failed to parse JSON)`, response);
      
      // Try to get text response for more info
      try {
        const textResponse = await response.clone().text();
        console.error(`API Error response text:`, textResponse);
      } catch (textError) {
        console.error(`Could not get response text:`, textError);
      }
      
      const error = new Error('API request failed');
      throw Object.assign(error, { status: response.status });
    }
  }

  try {
    console.log(`API Client: Parsing JSON from successful response`);
    const data = await response.json();
    console.log(`API Client: Parsed JSON data:`, data);
    return data;
  } catch (jsonError) {
    console.log('API Client: Response contains no JSON, returning empty success object');
    
    // Try to get text response for debugging
    try {
      const textResponse = await response.clone().text();
      console.log(`API Client: Non-JSON response text:`, textResponse);
    } catch (textError) {
      console.error(`API Client: Could not get response text:`, textError);
    }
    
    return { success: true };
  }
};

// Auth endpoints
export const authAPI = {
  login: async (email: string, password: string) => {
    const response = await fetch(`${API_BASE_URL}/auth/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ email, password }),
    });
    return handleResponse(response);
  },
  
  register: async (name: string, email: string, password: string) => {
    const response = await fetch(`${API_BASE_URL}/auth/register`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ name, email, password }),
    });
    return handleResponse(response);
  },
};

// User endpoints
export const userAPI = {
  getProfile: async () => {
    const response = await fetch(`${API_BASE_URL}/users/profile`, {
      headers: getAuthHeaders(),
    });
    return handleResponse(response);
  },
  
  updateName: async (name: string) => {
    const response = await fetch(`${API_BASE_URL}/users/name`, {
      method: 'PUT',
      headers: getAuthHeaders(),
      body: JSON.stringify({ name }),
    });
    return handleResponse(response);
  },
  
  updateEmail: async (email: string) => {
    const response = await fetch(`${API_BASE_URL}/users/email`, {
      method: 'PUT',
      headers: getAuthHeaders(),
      body: JSON.stringify({ email }),
    });
    return handleResponse(response);
  },
  
  updatePassword: async (oldPassword: string, newPassword: string) => {
    const response = await fetch(`${API_BASE_URL}/users/password`, {
      method: 'PUT',
      headers: getAuthHeaders(),
      body: JSON.stringify({ old_password: oldPassword, new_password: newPassword }),
    });
    return handleResponse(response);
  },
};

// Admin (root) user endpoints
export const adminAPI = {
  getUsers: async (page = 1, perPage = 10) => {
    const response = await fetch(`${API_BASE_URL}/users/root?page=${page}&per_page=${perPage}`, {
      headers: getAuthHeaders(),
    });
    return handleResponse(response);
  },
  
  createUser: async (userData: { name: string, email: string, password: string, role: number }) => {
    const response = await fetch(`${API_BASE_URL}/users/root`, {
      method: 'POST',
      headers: getAuthHeaders(),
      body: JSON.stringify(userData),
    });
    return handleResponse(response);
  },
  
  deleteUser: async (userId: string) => {
    const response = await fetch(`${API_BASE_URL}/users/root/${userId}`, {
      method: 'DELETE',
      headers: getAuthHeaders(),
    });
    return handleResponse(response);
  },
};

// Article endpoints
export const articleAPI = {
  listArticles: async (page = 1, perPage = 10, filters?: { is_read?: boolean, is_starred?: boolean }) => {
    let url = `${API_BASE_URL}/articles?page=${page}&per_page=${perPage}`;
    
    if (filters) {
      console.log("API client: received filters:", filters);
      
      // Handle filter conditions (is_read and is_starred)
      // API parameter meaning:
      // is_read=true: Get read articles
      // is_read=false: Get unread articles
      if (filters.is_read !== undefined) {
        url += `&is_read=${filters.is_read}`;
        console.log("API client: adding is_read=" + filters.is_read + 
                   " to URL, which will return " + (filters.is_read ? "read" : "unread") + " articles");
      } else {
        console.log("API client: is_read filter not provided, showing all articles (read status)");
      }
      
      // API parameter meaning:
      // is_starred=true: Get starred articles
      // is_starred=false: Get non-starred articles
      if (filters.is_starred !== undefined) {
        url += `&is_starred=${filters.is_starred}`;
        console.log("API client: adding is_starred=" + filters.is_starred + 
                   " to URL, which will return " + (filters.is_starred ? "starred" : "non-starred") + " articles");
      } else {
        console.log("API client: is_starred filter not provided, showing all articles (starred status)");
      }
    } else {
      console.log("API client: no filters provided, showing all articles");
    }
    
    console.log("API client: final URL:", url);
    
    const response = await fetch(url, {
      headers: getAuthHeaders(),
    });
    return handleResponse(response);
  },
  
  getArticle: async (id: string) => {
    const response = await fetch(`${API_BASE_URL}/articles/${id}`, {
      headers: getAuthHeaders(),
    });
    return handleResponse(response);
  },
  
  createArticle: async (url: string, categoryName?: string, tags?: string[]) => {
    const response = await fetch(`${API_BASE_URL}/articles`, {
      method: 'POST',
      headers: getAuthHeaders(),
      body: JSON.stringify({ 
        url, 
        category_name: categoryName,
        tags 
      }),
    });
    return handleResponse(response);
  },
  
  deleteArticle: async (id: string) => {
    const response = await fetch(`${API_BASE_URL}/articles/${id}`, {
      method: 'DELETE',
      headers: getAuthHeaders(),
    });
    // Handle 204 No Content response
    if (response.status === 204) return { success: true };
    return handleResponse(response);
  },
  
  updateArticleStatus: async (id: string, status: { is_read?: boolean, is_starred?: boolean }) => {
    console.log(`API Client: Updating article ${id} status:`, status);
    
    // Verify parameters
    if (!id) {
      console.error('API Client: Missing article ID for updateArticleStatus');
      throw new Error('Missing article ID');
    }

    // Validate status object
    if (status.is_read === undefined && status.is_starred === undefined) {
      console.error('API Client: Invalid status object, both is_read and is_starred are undefined');
      throw new Error('Invalid status object, must specify at least one of is_read or is_starred');
    }
    
    try {
      // Log request details
      console.log(`API Client: Sending PATCH request to ${API_BASE_URL}/articles/${id} with body:`, JSON.stringify(status));
      console.log('API Client: Using headers:', getAuthHeaders());
      
      const response = await fetch(`${API_BASE_URL}/articles/${id}`, {
        method: 'PATCH',
        headers: getAuthHeaders(),
        body: JSON.stringify(status),
      });
      
      console.log(`API Client: Raw status update response:`, {
        status: response.status,
        statusText: response.statusText,
        ok: response.ok,
        url: response.url,
        headers: Object.fromEntries(response.headers.entries()) // Log headers as an object
      });

      let responseBody = null;
      try {
        const clonedResponse = response.clone();
        const contentType = clonedResponse.headers.get("content-type");
        if (contentType && contentType.includes("application/json")) {
          responseBody = await clonedResponse.json();
          console.log(`API Client: Parsed JSON response body:`, responseBody);
        } else {
          responseBody = await clonedResponse.text();
          console.log(`API Client: Received non-JSON response body (text):`, responseBody);
        }
      } catch (e) {
        console.error(`API Client: Error parsing response body:`, e);
        try {
          const textBody = await response.text(); 
          console.log(`API Client: Fallback text response body after parse error:`, textBody);
        } catch (textErr) {
          console.error(`API Client: Error reading text body after JSON parse error:`, textErr);
        }
      }
      
      // const result = await handleResponse(response); // Temporarily commented out
      // console.log(`API Client: Status update result (from handleResponse):`, result); // Temporarily commented out
      
      if (response.ok) {
        if (responseBody && typeof responseBody === 'object' && responseBody !== null && 'data' in responseBody) {
             console.log(`API Client: Returning responseBody.data due to response.ok and 'data' field presence.`);
             return (responseBody as any).data; 
        } else if (responseBody) {
            console.log(`API Client: Returning full responseBody due to response.ok but no 'data' field or non-object.`);
            return responseBody;
        }
        console.log(`API Client: Returning {success: true} due to response.ok and no parsable body with 'data'.`);
        return { success: true }; 
    } else {
        console.error(`API Client: Response not OK (${response.status}). Throwing error with body:`, responseBody);
        throw new Error(`API request failed with status ${response.status}. Body: ${JSON.stringify(responseBody)}`);
    }

  } catch (error) {
    console.error(`API Client: Error updating article status (outer catch):`, error);
    throw error; 
  }
},
  
  updateArticleCategory: async (id: string, category: string) => {
    console.log(`API Client: Updating article ${id} category to:`, category);
    
    try {
      const response = await fetch(`${API_BASE_URL}/articles/${id}/category`, {
        method: 'PATCH',
        headers: getAuthHeaders(),
        body: JSON.stringify({ category }),
      });
      
      console.log(`API Client: Category update response:`, {
        status: response.status,
        statusText: response.statusText,
        ok: response.ok
      });
      
      const result = await handleResponse(response);
      console.log(`API Client: Category update result:`, result);
      return result;
    } catch (error) {
      console.error(`API Client: Error updating article category:`, error);
      throw error;
    }
  },
  
  rescrapeArticle: async (id: string) => {
    const response = await fetch(`${API_BASE_URL}/articles/${id}/rescrape`, {
      method: 'POST',
      headers: getAuthHeaders(),
    });
    return handleResponse(response);
  },
  
  addTags: async (id: string, tags: string[]) => {
    const response = await fetch(`${API_BASE_URL}/articles/${id}/tags`, {
      method: 'POST',
      headers: getAuthHeaders(),
      body: JSON.stringify({ tags }),
    });
    return handleResponse(response);
  },
  
  removeTags: async (id: string, tags: string[]) => {
    const response = await fetch(`${API_BASE_URL}/articles/${id}/tags`, {
      method: 'DELETE',
      headers: getAuthHeaders(),
      body: JSON.stringify({ tags }),
    });
    return handleResponse(response);
  },
  
  searchArticles: async (query: string, page = 1, perPage = 10, filters?: { is_read?: boolean, is_starred?: boolean }) => {
    let url = `${API_BASE_URL}/articles/search?q=${encodeURIComponent(query)}&page=${page}&per_page=${perPage}`;
    
    if (filters) {
      if (filters.is_read !== undefined) url += `&is_read=${filters.is_read}`;
      if (filters.is_starred !== undefined) url += `&is_starred=${filters.is_starred}`;
    }
    
    const response = await fetch(url, {
      headers: getAuthHeaders(),
    });
    return handleResponse(response);
  },
};

// Category endpoints
export const categoryAPI = {
  listCategories: async (page = 1, perPage = 10) => {
    const response = await fetch(`${API_BASE_URL}/categories?page=${page}&per_page=${perPage}`, {
      headers: getAuthHeaders(),
    });
    return handleResponse(response);
  },
  
  createCategory: async (name: string) => {
    const response = await fetch(`${API_BASE_URL}/categories`, {
      method: 'POST',
      headers: getAuthHeaders(),
      body: JSON.stringify({ name }),
    });
    return handleResponse(response);
  },
  
  deleteCategory: async (id: string) => {
    const response = await fetch(`${API_BASE_URL}/categories/${id}`, {
      method: 'DELETE',
      headers: getAuthHeaders(),
    });
    return handleResponse(response);
  },
  
  getArticlesByCategory: async (identifier: string, page = 1, perPage = 10) => {
    const response = await fetch(`${API_BASE_URL}/categories/${identifier}/articles?page=${page}&per_page=${perPage}`, {
      headers: getAuthHeaders(),
    });
    return handleResponse(response);
  },
  
  searchCategories: async (query: string, page = 1, perPage = 10) => {
    const response = await fetch(`${API_BASE_URL}/categories/search?q=${encodeURIComponent(query)}&page=${page}&per_page=${perPage}`, {
      headers: getAuthHeaders(),
    });
    return handleResponse(response);
  },
};

// Tag endpoints
export const tagAPI = {
  listTags: async (page = 1, perPage = 10) => {
    const response = await fetch(`${API_BASE_URL}/tags?page=${page}&per_page=${perPage}`, {
      headers: getAuthHeaders(),
    });
    return handleResponse(response);
  },
  
  createTag: async (name: string) => {
    const response = await fetch(`${API_BASE_URL}/tags`, {
      method: 'POST',
      headers: getAuthHeaders(),
      body: JSON.stringify({ name }),
    });
    return handleResponse(response);
  },
  
  deleteTag: async (id: string) => {
    const response = await fetch(`${API_BASE_URL}/tags/${id}`, {
      method: 'DELETE',
      headers: getAuthHeaders(),
    });
    return handleResponse(response);
  },
  
  getArticlesByTag: async (identifier: string, page = 1, perPage = 10) => {
    const response = await fetch(`${API_BASE_URL}/tags/${identifier}/articles?page=${page}&per_page=${perPage}`, {
      headers: getAuthHeaders(),
    });
    return handleResponse(response);
  },
  
  searchTags: async (query: string, page = 1, perPage = 10) => {
    const response = await fetch(`${API_BASE_URL}/tags/search?q=${encodeURIComponent(query)}&page=${page}&per_page=${perPage}`, {
      headers: getAuthHeaders(),
    });
    return handleResponse(response);
  },
};

// App config endpoints
export const configAPI = {
  getConfig: async () => {
    const response = await fetch(`${API_BASE_URL}/config`, {
      headers: getAuthHeaders(),
    });
    return handleResponse(response);
  },
  
  updateConfig: async (config: {
    bypass_refer?: boolean;
    custom_scrape_retry_times?: number;
    custom_scrape_timeout_seconds?: number;
    custom_user_agent?: string;
    custom_user_proxy?: string;
    extract_links?: boolean;
    llm_auto_gen_abstract?: boolean;
    llm_auto_gen_tags?: boolean;
    llm_profile_id?: string;
    llm_provider?: string;
    locale?: string;
    scrape_img_offline?: boolean;
  }) => {
    const response = await fetch(`${API_BASE_URL}/config`, {
      method: 'PUT',
      headers: getAuthHeaders(),
      body: JSON.stringify(config),
    });
    return handleResponse(response);
  },
};

// Root-only operations for admin
export const rootAPI = {
  // Category management
  listUserCategories: async (targetUserId: string, page = 1, perPage = 10) => {
    const response = await fetch(`${API_BASE_URL}/users/root/categories?target_user_id=${targetUserId}&page=${page}&per_page=${perPage}`, {
      headers: getAuthHeaders(),
    });
    return handleResponse(response);
  },
  
  createUserCategory: async (targetUserId: string, name: string) => {
    const response = await fetch(`${API_BASE_URL}/users/root/categories`, {
      method: 'POST',
      headers: getAuthHeaders(),
      body: JSON.stringify({ target_user_id: targetUserId, name }),
    });
    return handleResponse(response);
  },
  
  // Tag management
  listUserTags: async (targetUserId: string, page = 1, perPage = 10) => {
    const response = await fetch(`${API_BASE_URL}/users/root/tags?target_user_id=${targetUserId}&page=${page}&per_page=${perPage}`, {
      headers: getAuthHeaders(),
    });
    return handleResponse(response);
  },
  
  createUserTag: async (targetUserId: string, name: string) => {
    const response = await fetch(`${API_BASE_URL}/users/root/tags`, {
      method: 'POST',
      headers: getAuthHeaders(),
      body: JSON.stringify({ target_user_id: targetUserId, name }),
    });
    return handleResponse(response);
  },
  
  // Article management
  getUserArticles: async (userId: string, page = 1, perPage = 10) => {
    const response = await fetch(`${API_BASE_URL}/users/root/articles/user?user_id=${userId}&page=${page}&per_page=${perPage}`, {
      headers: getAuthHeaders(),
    });
    return handleResponse(response);
  },
  
  searchAllArticles: async (query: string, userId: string | null = null, page = 1, perPage = 10) => {
    let url = `${API_BASE_URL}/users/root/articles/search?page=${page}&per_page=${perPage}`;
    
    if (query) url += `&q=${encodeURIComponent(query)}`;
    if (userId) url += `&user_id=${userId}`;
    
    const response = await fetch(url, {
      headers: getAuthHeaders(),
    });
    return handleResponse(response);
  },
  
  // User config management
  listAllConfigs: async (page = 1, perPage = 10) => {
    const response = await fetch(`${API_BASE_URL}/users/root/config?page=${page}&per_page=${perPage}`, {
      headers: getAuthHeaders(),
    });
    return handleResponse(response);
  },
  
  getUserConfig: async (userId: string) => {
    const response = await fetch(`${API_BASE_URL}/users/root/config/${userId}`, {
      headers: getAuthHeaders(),
    });
    return handleResponse(response);
  },
  
  updateUserConfig: async (userId: string, config: any) => {
    const response = await fetch(`${API_BASE_URL}/users/root/config/${userId}`, {
      method: 'PUT',
      headers: getAuthHeaders(),
      body: JSON.stringify(config),
    });
    return handleResponse(response);
  },
};

// Default export with all API namespaces
export default {
  auth: authAPI,
  user: userAPI,
  admin: adminAPI,
  article: articleAPI,
  category: categoryAPI,
  tag: tagAPI,
  config: configAPI,
  root: rootAPI,
}; 