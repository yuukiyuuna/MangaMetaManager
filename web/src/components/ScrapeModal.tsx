import React, { useState, useEffect } from 'react';
import axios from 'axios';
import { Search, X, Check, Loader2, Calendar, ArrowRight } from 'lucide-react';
import { showToast } from './Toast';

interface ScrapeModalProps {
  itemId: number;
  itemType: 'series' | 'book';
  initialTitle: string;
  onClose: () => void;
  onScraped: () => void;
}

interface Provider {
  id: string;
  name: string;
}

interface SearchResult {
  id: string;
  title: string;
  series: string;
  coverUrl: string;
  releaseDate?: string;
}

const ScrapeModal: React.FC<ScrapeModalProps> = ({ itemId, itemType, initialTitle, onClose, onScraped }) => {
  const [providers, setProviders] = useState<Provider[]>([]);
  const [selectedProvider, setSelectedProvider] = useState('');
  const [query, setQuery] = useState(initialTitle);
  const [results, setResults] = useState<SearchResult[]>([]);
  const [loading, setLoading] = useState(false);
  
  // Preview mode states
  const [previewMode, setPreviewMode] = useState(false);
  const [oldData, setOldData] = useState<any>(null);
  const [newData, setNewData] = useState<any>(null);
  const [mergedData, setMergedData] = useState<any>(null);
  const [saving, setSaving] = useState(false);

  const fieldsToCompare = itemType === 'series' 
    ? ['title', 'originalTitle', 'series', 'alternateSeries', 'author', 'translator', 'publisher', 'genre', 'tags', 'summary', 'year', 'month', 'day', 'web', 'language', 'gtin', 'type', 'ageRating']
    : ['title', 'originalTitle', 'author', 'translator', 'publisher', 'genre', 'tags', 'summary', 'year', 'month', 'day', 'web', 'language', 'gtin', 'pageCount', 'type', 'ageRating'];

  useEffect(() => {
    const fetchProviders = async () => {
      try {
        const res = await axios.get('/api/metadata/providers');
        setProviders(res.data);
        if (res.data.length > 0) setSelectedProvider(res.data[0].id);
      } catch (err) {
        console.error(err);
      }
    };
    fetchProviders();
  }, []);

  const handleSearch = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    try {
      const res = await axios.get(`/api/metadata/providers/${selectedProvider}/search?q=${encodeURIComponent(query)}`);
      setResults(res.data);
    } catch (err) {
      console.error(err);
      showToast('Search failed: ' + (err as any).message, 'error');
    }
    setLoading(false);
  };

  const handlePreparePreview = async (metadataId: string) => {
    setLoading(true);
    try {
      const [detailsRes, currentRes] = await Promise.all([
        axios.get(`/api/metadata/providers/${selectedProvider}/details/${metadataId}`),
        axios.get(itemType === 'series' ? `/api/manga/${itemId}` : `/api/manga/books/${itemId}`)
      ]);
      
      const oldD = {
        title: currentRes.data.title || '',
        originalTitle: currentRes.data.originalTitle || '',
        series: currentRes.data.series || '',
        alternateSeries: currentRes.data.alternateSeries || '',
        author: currentRes.data.author || '',
        translator: currentRes.data.translator || '',
        publisher: currentRes.data.publisher || '',
        genre: currentRes.data.genre || '',
        tags: currentRes.data.tags || '',
        summary: currentRes.data.summary || '',
        year: currentRes.data.year || 0,
        month: currentRes.data.month || 0,
        day: currentRes.data.day || 0,
        web: currentRes.data.web || '',
        language: currentRes.data.language || '',
        gtin: currentRes.data.gtin || '',
        pageCount: currentRes.data.pageCount || 0,
        type: currentRes.data.type || '漫画',
        ageRating: currentRes.data.ageRating || ''
      };
      
      const newD = {
        title: detailsRes.data.Title || '',
        originalTitle: detailsRes.data.OriginalTitle || '',
        series: detailsRes.data.Series || '',
        alternateSeries: detailsRes.data.AlternateSeries || '',
        author: detailsRes.data.Writer || '',
        translator: detailsRes.data.Translator || '',
        publisher: detailsRes.data.Publisher || '',
        genre: detailsRes.data.Genre || '',
        tags: detailsRes.data.Tags || '',
        summary: detailsRes.data.Summary || '',
        year: detailsRes.data.Year || 0,
        month: detailsRes.data.Month || 0,
        day: detailsRes.data.Day || 0,
        web: detailsRes.data.Web || '',
        language: detailsRes.data.LanguageISO || '',
        gtin: detailsRes.data.GTIN || '',
        pageCount: detailsRes.data.PageCount || 0,
        type: (detailsRes.data.Manga === 'No' ? '小说' : '漫画'),
        ageRating: detailsRes.data.AgeRating || ''
      };
      
      setNewData(newD);
      setOldData(oldD);

      // Smart Merge Initialization:
      // If old field is empty, automatically use the new value.
      // If old field has value, keep it and wait for user to click the arrow to replace.
      const initialMerged: any = { ...oldD };
      Object.keys(newD).forEach(key => {
        if (!initialMerged[key] && (newD as any)[key]) {
          initialMerged[key] = (newD as any)[key];
        }
      });

      setMergedData(initialMerged); 
      setPreviewMode(true);
    } catch (err) {
      console.error(err);
      showToast('Failed to load details for comparison', 'error');
    }
    setLoading(false);
  };

  const applyNewValue = (field: string) => {
    setMergedData({ ...mergedData, [field]: newData[field] });
  };

  const handleConfirmSave = async () => {
    setSaving(true);
    try {
      const endpoint = itemType === 'series' 
        ? `/api/manga/${itemId}/scrape` 
        : `/api/manga/books/${itemId}/scrape`;
        
      await axios.post(endpoint, mergedData);
      showToast('Metadata successfully updated and written to archive!');
      onScraped();
      onClose();
    } catch (err) {
      console.error(err);
      showToast('Save failed: ' + (err as any).message, 'error');
    }
    setSaving(false);
  };

  if (previewMode && oldData && newData) {
    return (
      <div className="fixed inset-0 bg-black/60 flex items-center justify-center z-50 p-4 backdrop-blur-sm">
        <div className="bg-white rounded-2xl w-full max-w-5xl max-h-[90vh] flex flex-col shadow-2xl overflow-hidden animate-in fade-in zoom-in duration-200">
          <div className="p-6 border-b flex justify-between items-center bg-gray-50/50">
            <div>
              <h3 className="text-xl font-bold text-gray-900">Compare & Merge Metadata</h3>
              <p className="text-sm text-gray-500 mt-0.5">Click the arrow to replace your current value with the newly scraped value.</p>
            </div>
            <button onClick={() => setPreviewMode(false)} className="p-2 hover:bg-gray-100 rounded-full transition-colors text-gray-400">
              <X className="w-6 h-6" />
            </button>
          </div>
          
          <div className="flex-1 overflow-y-auto p-6">
            <table className="w-full text-left border-collapse">
              <thead>
                <tr className="bg-gray-50 border-b border-gray-200">
                  <th className="px-4 py-3 text-xs font-bold text-gray-500 uppercase w-32">Field</th>
                  <th className="px-4 py-3 text-xs font-bold text-gray-500 uppercase w-1/2 border-r">Current Value (Will be saved)</th>
                  <th className="px-4 py-3 text-xs font-bold text-gray-500 uppercase w-10 text-center"></th>
                  <th className="px-4 py-3 text-xs font-bold text-gray-500 uppercase w-1/2">Scraped Value</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-100">
                {fieldsToCompare.map((field) => {
                  const oVal = mergedData[field];
                  const nVal = newData[field];
                  if (oVal === nVal || !nVal) return null; // "一样不展示" or no new data
                  
                  return (
                    <tr key={field} className="hover:bg-gray-50/50">
                      <td className="px-4 py-4 text-xs font-bold text-gray-500 uppercase tracking-wider">{field}</td>
                      <td className="px-4 py-4 border-r">
                        <div className={`text-sm ${oVal ? 'text-gray-900' : 'text-gray-400 italic'}`}>{oVal || 'Empty'}</div>
                      </td>
                      <td className="px-2 py-4 text-center">
                        <button 
                          onClick={() => applyNewValue(field)}
                          className="p-2 bg-blue-50 text-blue-600 rounded-full hover:bg-blue-600 hover:text-white transition-all shadow-sm group"
                          title="Apply Scraped Value"
                        >
                          <ArrowRight className="w-4 h-4 group-hover:-translate-x-0.5 transition-transform" />
                        </button>
                      </td>
                      <td className="px-4 py-4">
                        <div className="text-sm text-blue-800 font-medium bg-blue-50/50 p-2 rounded">{nVal}</div>
                      </td>
                    </tr>
                  )
                })}
              </tbody>
            </table>
            
            {fieldsToCompare.every(f => mergedData[f] === newData[f] || !newData[f]) && (
              <div className="text-center py-12 text-gray-500 italic">
                No new differences to apply.
              </div>
            )}
          </div>
          
          <div className="p-6 bg-gray-50 border-t flex justify-end gap-3">
            <button
              onClick={() => setPreviewMode(false)}
              className="px-6 py-2.5 bg-white border border-gray-300 rounded-xl text-sm font-bold text-gray-700 hover:bg-gray-100 transition-all shadow-sm"
            >
              Back to Search
            </button>
            <button
              onClick={handleConfirmSave}
              disabled={saving}
              className="px-6 py-2.5 bg-blue-600 text-white rounded-xl text-sm font-bold hover:bg-blue-700 transition-all shadow-md shadow-blue-100 flex items-center disabled:opacity-50"
            >
              {saving ? <Loader2 className="w-4 h-4 animate-spin mr-2" /> : <Check className="w-4 h-4 mr-2" />}
              Confirm & Save
            </button>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="fixed inset-0 bg-black/60 flex items-center justify-center z-50 p-4 backdrop-blur-sm">
      <div className="bg-white rounded-2xl w-full max-w-4xl max-h-[90vh] flex flex-col shadow-2xl overflow-hidden animate-in fade-in zoom-in duration-200">
        <div className="p-6 border-b flex justify-between items-center bg-gray-50/50">
          <div>
            <h3 className="text-xl font-bold text-gray-900">Scrape {itemType === 'series' ? 'Series' : 'Book'}</h3>
            <p className="text-sm text-gray-500 mt-0.5">Find metadata for: <span className="font-medium text-gray-700">{initialTitle}</span></p>
          </div>
          <button onClick={onClose} className="p-2 hover:bg-gray-100 rounded-full transition-colors text-gray-400">
            <X className="w-6 h-6" />
          </button>
        </div>

        <div className="p-6 bg-white border-b shadow-inner">
          <form onSubmit={handleSearch} className="flex gap-4">
            <select
              value={selectedProvider}
              onChange={(e) => setSelectedProvider(e.target.value)}
              className="bg-gray-50 border border-gray-200 rounded-xl px-4 py-2.5 focus:ring-2 focus:ring-blue-500 outline-none transition-all font-medium text-sm"
            >
              {providers.map((p) => (
                <option key={p.id} value={p.id}>{p.name}</option>
              ))}
            </select>
            <div className="flex-1 relative">
              <input
                type="text"
                value={query}
                onChange={(e) => setQuery(e.target.value)}
                className="w-full border border-gray-200 rounded-xl pl-10 pr-4 py-2.5 focus:ring-2 focus:ring-blue-500 outline-none transition-all text-sm"
                placeholder="Search title..."
              />
              <Search className="absolute left-3.5 top-3 text-gray-400 w-4 h-4" />
            </div>
            <button
              type="submit"
              disabled={loading}
              className="bg-blue-600 text-white px-6 py-2.5 rounded-xl hover:bg-blue-700 disabled:opacity-50 transition-all flex items-center shadow-md shadow-blue-100 font-bold text-sm"
            >
              {loading ? <Loader2 className="w-4 h-4 animate-spin" /> : <Search className="w-4 h-4 mr-2" />}
              {loading ? 'Searching...' : 'Search'}
            </button>
          </form>
        </div>

        <div className="flex-1 overflow-y-auto p-6 bg-gray-50/30">
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-2 gap-5">
            {results.map((res) => (
              <div key={res.id} className="flex border border-gray-100 rounded-2xl p-4 hover:border-blue-400 hover:shadow-lg transition-all bg-white group cursor-default">
                <div className="w-24 h-32 bg-gray-100 rounded-lg overflow-hidden flex-shrink-0 shadow-sm border border-gray-50">
                  {res.coverUrl ? (
                    <img 
                      src={res.coverUrl} 
                      alt="" 
                      className="w-full h-full object-cover transition-transform group-hover:scale-110" 
                      referrerPolicy="no-referrer"
                    />
                  ) : (
                    <div className="w-full h-full flex items-center justify-center text-gray-300">
                      <Search className="w-8 h-8 opacity-20" />
                    </div>
                  )}
                </div>
                <div className="ml-5 flex-1 min-w-0 flex flex-col justify-between">
                  <div>
                    <h4 className="font-bold text-gray-900 leading-tight group-hover:text-blue-600 transition-colors line-clamp-2" title={res.title}>{res.title}</h4>
                    <p className="text-[10px] text-gray-400 mt-1.5 font-black uppercase tracking-wider">{res.series || 'No Series'}</p>
                    {res.releaseDate && (
                      <div className="flex items-center text-[10px] text-gray-400 mt-2 bg-gray-50 px-2 py-1 rounded w-fit">
                        <Calendar className="w-3 h-3 mr-1" />
                        {res.releaseDate}
                      </div>
                    )}
                  </div>
                  <button
                    onClick={() => handlePreparePreview(res.id)}
                    disabled={loading}
                    className="w-full mt-3 bg-blue-50 hover:bg-blue-600 hover:text-white text-blue-600 py-2 rounded-xl text-xs font-bold transition-all flex items-center justify-center border border-blue-100 group-hover:border-blue-600"
                  >
                    {loading ? <Loader2 className="w-3 h-3 animate-spin mr-1" /> : <Search className="w-3 h-3 mr-1" />}
                    Preview Details
                  </button>
                </div>
              </div>
            ))}
            {results.length === 0 && !loading && (
              <div className="col-span-full text-center py-20">
                <div className="mx-auto w-16 h-16 bg-gray-100 rounded-full flex items-center justify-center mb-4">
                  <Search className="w-8 h-8 text-gray-300" />
                </div>
                <p className="text-gray-400 font-medium italic">Enter a title and search for metadata</p>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
};

export default ScrapeModal;
