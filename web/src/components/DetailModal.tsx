import React, { useState, useEffect } from 'react';
import axios from 'axios';
import { X, Book, Database, Edit3, Save, Code, Loader2, Link2, Hash, FileText, RefreshCw } from 'lucide-react';
import { showToast } from './Toast';
import Editor from '@monaco-editor/react';

interface DetailModalProps {
  itemId: number;
  itemType: 'series' | 'book';
  onClose: () => void;
  onSaved: () => void;
}

const DetailModal: React.FC<DetailModalProps> = ({ itemId, itemType, onClose, onSaved }) => {
  const [activeTab, setActiveTab] = useState<'form' | 'xml'>('form');
  
  // Data State
  const [loading, setLoading] = useState(true);
  const [formData, setFormData] = useState<any>(null);
  const [savingForm, setSavingForm] = useState(false);

  // XML State
  const [xmlContent, setXmlContent] = useState('');
  const [loadingXml, setLoadingXml] = useState(false);
  const [savingXml, setSavingXml] = useState(false);

  const fetchLatestData = async () => {
    console.log(`[DetailModal] Fetching latest data for ${itemType} ID: ${itemId}`);
    setLoading(true);
    try {
      const endpoint = itemType === 'series' ? `/api/manga/${itemId}` : `/api/manga/books/${itemId}`;
      const res = await axios.get(endpoint);
      console.log(`[DetailModal] Received data:`, res.data);
      setFormData(res.data);
    } catch (err) {
      console.error('[DetailModal] Fetch error:', err);
      showToast('Failed to fetch latest data from database', 'error');
    }
    setLoading(false);
  };

  const loadXml = async () => {
    setLoadingXml(true);
    try {
      const endpoint = itemType === 'series' ? `/api/manga/${itemId}/xml` : `/api/manga/books/${itemId}/xml`;
      const res = await axios.get(endpoint, { responseType: 'text' });
      setXmlContent(res.data);
    } catch (err) {
      console.error(err);
      showToast('Failed to load XML.', 'error');
      setXmlContent('<?xml version="1.0" encoding="utf-8"?>\n<ComicInfo xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xmlns:xsd="http://www.w3.org/2001/XMLSchema">\n</ComicInfo>');
    }
    setLoadingXml(false);
  };

  const handleSaveForm = async () => {
    setSavingForm(true);
    try {
      const endpoint = itemType === 'series' ? `/api/manga/${itemId}` : `/api/manga/books/${itemId}`;
      await axios.patch(endpoint, formData);
      showToast('Metadata saved successfully!');
      onSaved();
    } catch (err) {
      console.error(err);
      showToast('Failed to save metadata.', 'error');
    }
    setSavingForm(false);
  };

  const handleSaveXml = async () => {
    setSavingXml(true);
    try {
      const endpoint = itemType === 'series' ? `/api/manga/${itemId}/xml` : `/api/manga/books/${itemId}/xml`;
      await axios.put(endpoint, xmlContent, {
        headers: { 'Content-Type': 'application/xml' }
      });
      showToast('Raw XML saved successfully!');
      await fetchLatestData();
      onSaved();
    } catch (err: any) {
      console.error(err);
      showToast(err.response?.data?.error || 'Failed to save XML.', 'error');
    }
    setSavingXml(false);
  };

  useEffect(() => {
    fetchLatestData();
  }, [itemId, itemType]);

  useEffect(() => {
    if (activeTab === 'xml') {
      loadXml();
    }
  }, [activeTab]);

  if (loading || !formData) {
    return (
      <div className="fixed inset-0 bg-black/60 flex items-center justify-center z-50 p-4 backdrop-blur-sm">
        <div className="bg-white rounded-2xl p-12 flex flex-col items-center">
          <Loader2 className="w-8 h-8 animate-spin text-blue-600 mb-4" />
          <p className="text-gray-500 font-bold">Loading Meta...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="fixed inset-0 bg-black/60 flex items-center justify-center z-50 p-4 backdrop-blur-sm">
      <div className="bg-white rounded-2xl w-full max-w-5xl h-[90vh] flex flex-col shadow-2xl overflow-hidden animate-in fade-in zoom-in duration-200">
        
        {/* Header */}
        <div className="p-6 border-b flex justify-between items-center bg-gray-50/50">
          <div className="flex items-center space-x-3">
            <div className="p-2 bg-blue-100 rounded-lg">
              {itemType === 'series' ? <Database className="w-5 h-5 text-blue-600" /> : <Book className="w-5 h-5 text-blue-600" />}
            </div>
            <div>
              <h3 className="text-xl font-bold text-gray-900 truncate max-w-xl">{formData.title}</h3>
              <p className="text-sm text-gray-500">{itemType === 'series' ? 'Manga Series' : 'Book File'}</p>
            </div>
          </div>
          <div className="flex items-center gap-4">
             <button 
               onClick={fetchLatestData}
               className="p-2 hover:bg-gray-100 rounded-full text-gray-400"
               title="Force Refresh"
             >
               <RefreshCw className={`w-4 h-4 ${loading ? 'animate-spin' : ''}`} />
             </button>
             <span className={`px-2.5 py-1 text-[10px] font-black uppercase tracking-wider rounded-lg shadow-sm ${
               formData.status === 'Scraped' ? 'bg-green-100 text-green-700 border border-green-200' : 'bg-yellow-50 text-yellow-700 border border-yellow-100'
             }`}>
               {formData.status || 'Unscraped'}
             </span>
             <button onClick={onClose} className="p-2 hover:bg-gray-100 rounded-full transition-colors text-gray-400">
               <X className="w-6 h-6" />
             </button>
          </div>
        </div>

        {/* Tabs */}
        <div className="flex px-6 border-b bg-gray-50">
          <button
            onClick={() => setActiveTab('form')}
            className={`px-4 py-3 text-sm font-bold border-b-2 flex items-center transition-colors ${
              activeTab === 'form' ? 'border-blue-600 text-blue-600' : 'border-transparent text-gray-500 hover:text-gray-700'
            }`}
          >
            <Edit3 className="w-4 h-4 mr-2" /> Detailed Form
          </button>
          {itemType === 'book' && (
            <button
              onClick={() => setActiveTab('xml')}
              className={`px-4 py-3 text-sm font-bold border-b-2 flex items-center transition-colors ${
                activeTab === 'xml' ? 'border-blue-600 text-blue-600' : 'border-transparent text-gray-500 hover:text-gray-700'
              }`}
            >
              <Code className="w-4 h-4 mr-2" /> Raw XML
            </button>
          )}
        </div>

        {/* Content */}
        <div className="flex-1 overflow-y-auto bg-gray-50/30 relative">
          
          {/* Form Tab */}
          {activeTab === 'form' && (
            <div className="p-8 space-y-6">
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                <div className="lg:col-span-2">
                  <label className="text-xs font-black text-gray-400 uppercase tracking-widest mb-2 block">Display Title</label>
                  <input
                    type="text"
                    value={formData.title || ''}
                    onChange={e => setFormData({...formData, title: e.target.value})}
                    className="w-full border border-gray-200 rounded-xl px-4 py-2.5 focus:ring-2 focus:ring-blue-500 outline-none bg-white font-bold"
                  />
                </div>
                
                <div>
                  <label className="text-xs font-black text-gray-400 uppercase tracking-widest mb-2 block">Original Title</label>
                  <input
                    type="text"
                    value={formData.originalTitle || ''}
                    onChange={e => setFormData({...formData, originalTitle: e.target.value})}
                    className="w-full border border-gray-200 rounded-xl px-4 py-2.5 focus:ring-2 focus:ring-blue-500 outline-none bg-white"
                  />
                </div>

                <div>
                  <label className="text-xs font-black text-gray-400 uppercase tracking-widest mb-2 block">Author / Writer</label>
                  <input
                    type="text"
                    value={formData.author || ''}
                    onChange={e => setFormData({...formData, author: e.target.value})}
                    className="w-full border border-gray-200 rounded-xl px-4 py-2.5 focus:ring-2 focus:ring-blue-500 outline-none bg-white"
                  />
                </div>

                <div>
                  <label className="text-xs font-black text-gray-400 uppercase tracking-widest mb-2 block">Publisher</label>
                  <input
                    type="text"
                    value={formData.publisher || ''}
                    onChange={e => setFormData({...formData, publisher: e.target.value})}
                    className="w-full border border-gray-200 rounded-xl px-4 py-2.5 focus:ring-2 focus:ring-blue-500 outline-none bg-white"
                  />
                </div>

                <div>
                  <label className="text-xs font-black text-gray-400 uppercase tracking-widest mb-2 block">Genre / Tags</label>
                  <input
                    type="text"
                    value={formData.genre || ''}
                    onChange={e => setFormData({...formData, genre: e.target.value})}
                    className="w-full border border-gray-200 rounded-xl px-4 py-2.5 focus:ring-2 focus:ring-blue-500 outline-none bg-white"
                  />
                </div>

                <div>
                  <label className="text-xs font-black text-gray-400 uppercase tracking-widest mb-2 block">Release Date (Y-M-D)</label>
                  <div className="flex gap-2">
                    <input
                      type="number"
                      placeholder="Year"
                      value={formData.year || ''}
                      onChange={e => setFormData({...formData, year: parseInt(e.target.value) || 0})}
                      className="w-full border border-gray-200 rounded-xl px-3 py-2.5 focus:ring-2 focus:ring-blue-500 outline-none bg-white"
                    />
                    <input
                      type="number"
                      placeholder="Month"
                      value={formData.month || ''}
                      onChange={e => setFormData({...formData, month: parseInt(e.target.value) || 0})}
                      className="w-20 border border-gray-200 rounded-xl px-3 py-2.5 focus:ring-2 focus:ring-blue-500 outline-none bg-white"
                    />
                    <input
                      type="number"
                      placeholder="Day"
                      value={formData.day || ''}
                      onChange={e => setFormData({...formData, day: parseInt(e.target.value) || 0})}
                      className="w-20 border border-gray-200 rounded-xl px-3 py-2.5 focus:ring-2 focus:ring-blue-500 outline-none bg-white"
                    />
                  </div>
                </div>

                {itemType === 'book' && (
                  <div>
                    <label className="text-xs font-black text-gray-400 uppercase tracking-widest mb-2 block flex items-center">
                      <Hash className="w-3 h-3 mr-1" /> Page Count
                    </label>
                    <input
                      type="number"
                      value={formData.pageCount || ''}
                      onChange={e => setFormData({...formData, pageCount: parseInt(e.target.value) || 0})}
                      className="w-full border border-gray-200 rounded-xl px-4 py-2.5 focus:ring-2 focus:ring-blue-500 outline-none bg-white"
                    />
                  </div>
                )}

                <div className="md:col-span-2">
                  <label className="text-xs font-black text-gray-400 uppercase tracking-widest mb-2 block flex items-center">
                    <Link2 className="w-3 h-3 mr-1" /> Web URL
                  </label>
                  <input
                    type="text"
                    value={formData.web || ''}
                    onChange={e => setFormData({...formData, web: e.target.value})}
                    className="w-full border border-gray-200 rounded-xl px-4 py-2.5 focus:ring-2 focus:ring-blue-500 outline-none bg-white"
                    placeholder="https://bgm.tv/subject/..."
                  />
                </div>
              </div>

              <div>
                <label className="text-xs font-black text-gray-400 uppercase tracking-widest mb-2 block">Summary / Description</label>
                <textarea
                  value={formData.summary || ''}
                  onChange={e => setFormData({...formData, summary: e.target.value})}
                  rows={8}
                  className="w-full border border-gray-200 rounded-xl px-4 py-4 focus:ring-2 focus:ring-blue-500 outline-none bg-white leading-relaxed resize-none text-sm"
                />
              </div>
              
              <div className="bg-white p-4 rounded-xl flex items-center border border-gray-100 shadow-sm">
                 <FileText className="w-4 h-4 text-gray-400 mr-2 shrink-0" />
                 <span className="text-[10px] font-mono text-gray-400 break-all truncate">{formData.path}</span>
              </div>
            </div>
          )}

          {/* XML Tab */}
          {activeTab === 'xml' && (
            <div className="h-full flex flex-col">
              {loadingXml ? (
                <div className="flex-1 flex flex-col items-center justify-center text-gray-400">
                  <Loader2 className="w-8 h-8 animate-spin mb-4" />
                  <p>Loading XML from archive...</p>
                </div>
              ) : (
                <div className="flex-1">
                  <Editor
                    height="100%"
                    defaultLanguage="xml"
                    value={xmlContent}
                    onChange={(val) => setXmlContent(val || '')}
                    theme="vs-light"
                    options={{
                      minimap: { enabled: false },
                      wordWrap: 'on',
                      formatOnPaste: true,
                      fontSize: 14,
                    }}
                  />
                </div>
              )}
            </div>
          )}
        </div>

        {/* Footer */}
        <div className="p-4 bg-gray-50 border-t flex justify-end gap-3">
          <button
            onClick={onClose}
            className="px-6 py-2.5 bg-white border border-gray-300 rounded-xl text-sm font-bold text-gray-700 hover:bg-gray-100 transition-all shadow-sm"
          >
            Cancel
          </button>
          {activeTab === 'form' ? (
            <button
              onClick={handleSaveForm}
              disabled={savingForm}
              className="px-8 py-2.5 bg-blue-600 text-white rounded-xl text-sm font-bold hover:bg-blue-700 transition-all shadow-md shadow-blue-100 flex items-center disabled:opacity-50"
            >
              {savingForm ? <Loader2 className="w-4 h-4 animate-spin mr-2" /> : <Save className="w-4 h-4 mr-2" />}
              Save to Database & Archive
            </button>
          ) : (
            <button
              onClick={handleSaveXml}
              disabled={savingXml || loadingXml}
              className="px-8 py-2.5 bg-purple-600 text-white rounded-xl text-sm font-bold hover:bg-purple-700 transition-all shadow-md shadow-purple-100 flex items-center disabled:opacity-50"
            >
              {savingXml ? <Loader2 className="w-4 h-4 animate-spin mr-2" /> : <Save className="w-4 h-4 mr-2" />}
              Update Archive ZIP
            </button>
          )}
        </div>
      </div>
    </div>
  );
};

export default DetailModal;
