import React, { useState, useEffect } from 'react';
import axios from 'axios';
import { X, Wand2, Loader2 } from 'lucide-react';
import { showToast } from './Toast';

interface AutoScrapeModalProps {
  seriesId: number;
  onClose: () => void;
  onComplete: () => void;
}

interface Provider {
  id: string;
  name: string;
}

const getErrorMessage = (err: unknown, fallback: string) => {
  if (err && typeof err === 'object' && 'response' in err) {
    const response = (err as { response?: { data?: { error?: string } } }).response;
    return response?.data?.error || fallback;
  }
  return fallback;
};

const AutoScrapeModal: React.FC<AutoScrapeModalProps> = ({ seriesId, onClose, onComplete }) => {
  const [providers, setProviders] = useState<Provider[]>([]);
  const [selectedProvider, setSelectedProvider] = useState('');
  const [loading, setLoading] = useState(false);
  const [options, setOptions] = useState({
    updateTitle: true,
    updateAuthor: true,
    updateTranslator: true,
    updateSummary: true,
    updatePublisher: true,
    updateGenre: true,
    updateDate: true,
    updateWeb: true,
    updateLanguage: true,
    updatePageCount: true,
    updateType: true,
    updateAgeRating: true,
    updateGtin: true,
  });

  const optionLabels: Record<string, string> = {
    updateTitle: 'Title / Original Title',
    updateAuthor: 'Author',
    updateTranslator: 'Translator',
    updateSummary: 'Summary',
    updatePublisher: 'Publisher',
    updateGenre: 'Genre / Tags',
    updateDate: 'Release Date',
    updateWeb: 'Web URL',
    updateLanguage: 'Language',
    updatePageCount: 'Page Count',
    updateType: 'Type',
    updateAgeRating: 'Age Rating',
    updateGtin: 'GTIN / ISBN',
  };

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

  const handleAutoScrape = async () => {
    setLoading(true);
    try {
      await axios.post(`/api/manga/${seriesId}/auto-scrape-books`, {
        providerId: selectedProvider,
        options: options
      });
      showToast('Auto Scrape started in the background!');
      onComplete();
      onClose();
    } catch (err) {
      console.error(err);
      showToast(getErrorMessage(err, 'Failed to start auto scrape.'), 'error');
    }
    setLoading(false);
  };

  return (
    <div className="fixed inset-0 bg-black/60 flex items-center justify-center z-50 p-4 backdrop-blur-sm">
      <div className="bg-white rounded-2xl w-full max-w-md shadow-2xl overflow-hidden animate-in fade-in zoom-in duration-200">
        <div className="p-6 border-b flex justify-between items-center bg-gray-50/50">
          <div>
            <h3 className="text-xl font-bold text-gray-900 flex items-center">
              <Wand2 className="w-5 h-5 mr-2 text-purple-600" /> Auto Scrape Series
            </h3>
            <p className="text-sm text-gray-500 mt-1">Automatically match and scrape all books in this series</p>
          </div>
          <button onClick={onClose} className="p-2 hover:bg-gray-100 rounded-full transition-colors text-gray-400">
            <X className="w-6 h-6" />
          </button>
        </div>

        <div className="p-6 space-y-6">
          <div>
            <label className="block text-sm font-bold text-gray-700 mb-2">Select Provider</label>
            <select
              value={selectedProvider}
              onChange={(e) => setSelectedProvider(e.target.value)}
              className="w-full bg-white border border-gray-300 rounded-xl px-4 py-2.5 focus:ring-2 focus:ring-purple-500 outline-none transition-all text-sm"
            >
              {providers.map((p) => (
                <option key={p.id} value={p.id}>{p.name}</option>
              ))}
            </select>
          </div>

          <div>
            <label className="block text-sm font-bold text-gray-700 mb-3">Fields to Update</label>
            <div className="space-y-3 bg-gray-50 p-4 rounded-xl border border-gray-100">
              {Object.entries(options).map(([key, val]) => (
                <label key={key} className="flex items-center space-x-3 cursor-pointer group">
                  <input 
                    type="checkbox" 
                    checked={val} 
                    onChange={(e) => setOptions({...options, [key]: e.target.checked})}
                    className="w-4 h-4 rounded border-gray-300 text-purple-600 focus:ring-purple-500 transition-colors cursor-pointer"
                  />
                  <span className="text-sm font-medium text-gray-700 group-hover:text-purple-600 transition-colors capitalize">
                    {optionLabels[key] || key.replace('update', '')}
                  </span>
                </label>
              ))}
            </div>
          </div>
        </div>

        <div className="p-6 bg-gray-50 border-t flex gap-3 justify-end">
          <button
            onClick={onClose}
            className="px-6 py-2.5 bg-white border border-gray-300 rounded-xl text-sm font-bold text-gray-700 hover:bg-gray-100 transition-all shadow-sm"
          >
            Cancel
          </button>
          <button
            onClick={handleAutoScrape}
            disabled={loading || !selectedProvider}
            className="px-6 py-2.5 bg-purple-600 text-white rounded-xl text-sm font-bold hover:bg-purple-700 transition-all shadow-md shadow-purple-200 flex items-center disabled:opacity-50"
          >
            {loading ? <Loader2 className="w-4 h-4 animate-spin mr-2" /> : <Wand2 className="w-4 h-4 mr-2" />}
            Start Auto Scrape
          </button>
        </div>
      </div>
    </div>
  );
};

export default AutoScrapeModal;
