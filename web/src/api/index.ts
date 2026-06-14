import axios from 'axios';

// --- Types ---

export interface ProxySettings {
  enabled: boolean;
  type: string;
  host: string;
  port: number;
  username: string;
  password?: string;
  noProxy: string;
  timeoutSeconds: number;
}

export interface AppSettings {
  backupBeforeFlatten: boolean;
}

export interface LibraryFolder {
  ID: number;
  path: string;
}

export interface MangaBook {
  ID: number;
  filename: string;
  title: string;
  author: string;
  translator: string;
  publisher: string;
  genre: string;
  tags: string;
  volume: number;
  year: number;
  month: number;
  day: number;
  web: string;
  language: string;
  pageCount: number;
  type: string;
  ageRating: string;
  status: string;
  summary: string;
  path: string;
  lastError: string;
}

export interface MangaSeries {
  ID: number;
  title: string;
  originalTitle: string;
  series: string;
  alternateSeries: string;
  path: string;
  author: string;
  translator: string;
  publisher: string;
  genre: string;
  tags: string;
  status: string;
  summary: string;
  year: number;
  month: number;
  day: number;
  web: string;
  language: string;
  type: string;
  ageRating: string;
  lastError: string;
  books: MangaBook[];
}

// --- API Methods ---

export const api = {
  // Settings
  getProxySettings: () => axios.get<ProxySettings>('/api/settings/proxy'),
  updateProxySettings: (data: ProxySettings) => axios.patch('/api/settings/proxy', data),
  testProxy: (testUrl: string = 'https://www.google.com') => 
    axios.post<{ success: boolean; error?: string }>('/api/settings/proxy/test', { testUrl }),
  
  getAppSettings: () => axios.get<AppSettings>('/api/settings/app'),
  updateAppSettings: (data: AppSettings) => axios.patch('/api/settings/app', data),

  // Library Folders
  getLibraryFolders: () => axios.get<LibraryFolder[]>('/api/library/folders'),
  addLibraryFolder: (path: string) => axios.post('/api/library/folders', { path }),
  removeLibraryFolder: (id: number) => axios.delete(`/api/library/folders/${id}`),
  scanLibrary: () => axios.post('/api/library/scan'),
  cleanLibrary: () => axios.post('/api/library/clean'),

  // Manga
  getSeriesList: (page: number = 1, size: number = 20) => 
    axios.get<MangaSeries[]>(`/api/manga?page=${page}&size=${size}`),
  getSeries: (id: number) => axios.get<MangaSeries>(`/api/manga/${id}`),
  updateSeries: (id: number, data: any) => axios.patch(`/api/manga/${id}`, data),
  deleteSeries: (id: number) => axios.delete(`/api/manga/${id}`),
  scrapeSeries: (id: number, data: any) => axios.post(`/api/manga/${id}/scrape`, data),
  autoScrapeBooks: (seriesId: number, providerId: string) => 
    axios.post(`/api/manga/${seriesId}/auto-scrape-books`, { providerId }),

  getBook: (id: number) => axios.get<MangaBook>(`/api/manga/books/${id}`),
  updateBook: (id: number, data: any) => axios.patch(`/api/manga/books/${id}`, data),
  scrapeBook: (id: number, data: any) => axios.post(`/api/manga/books/${id}/scrape`, data),

  // XML
  getSeriesXML: (id: number) => axios.get(`/api/manga/${id}/xml`, { responseType: 'text' }),
  updateSeriesXML: (id: number, xml: string) => 
    axios.put(`/api/manga/${id}/xml`, xml, { headers: { 'Content-Type': 'application/xml' } }),
  getBookXML: (id: number) => axios.get(`/api/manga/books/${id}/xml`, { responseType: 'text' }),
  updateBookXML: (id: number, xml: string) => 
    axios.put(`/api/manga/books/${id}/xml`, xml, { headers: { 'Content-Type': 'application/xml' } }),

  // Metadata Providers
  getProviders: () => axios.get<{ id: string; name: string }[]>('/api/metadata/providers'),
  searchProvider: (providerId: string, q: string) => 
    axios.get<any[]>(`/api/metadata/providers/${providerId}/search?q=${encodeURIComponent(q)}`),
  getProviderDetails: (providerId: string, metadataId: string) => 
    axios.get<any>(`/api/metadata/providers/${providerId}/details/${metadataId}`),
};
