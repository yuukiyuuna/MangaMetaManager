import React, { useEffect, useState, useMemo } from 'react';
import axios from 'axios';
import { RefreshCw, FolderOpen, Search, ChevronDown, ChevronRight, Book, ArrowUp, ArrowDown, LayoutList, Eye, Wand2, Eraser, Trash2 } from 'lucide-react';
import ScrapeModal from '../components/ScrapeModal';
import DetailModal from '../components/DetailModal';
import AutoScrapeModal from '../components/AutoScrapeModal';
import ConfirmModal from '../components/ConfirmModal';

import { showToast } from '../components/Toast';

interface BookItem {
  ID: number;
  filename: string;
  title: string;
  author: string;
  volume: number;
  status: string;
  summary: string;
  path: string;
}

interface Series {
  ID: number;
  title: string;
  path: string;
  author: string;
  genre: string;
  status: string;
  summary: string;
  books: BookItem[];
}

type SortField = 'title' | 'author' | 'status' | 'bookCount';
type SortOrder = 'asc' | 'desc';

const SortIcon = ({ field, sortField, sortOrder }: { field: SortField, sortField: SortField, sortOrder: SortOrder }) => {
  if (sortField !== field) return null;
  return sortOrder === 'asc' ? <ArrowUp className="ml-1 w-4 h-4 text-blue-500" /> : <ArrowDown className="ml-1 w-4 h-4 text-blue-500" />;
};

const LibraryPage: React.FC = () => {
  const [seriesList, setSeriesList] = useState<Series[]>([]);
  const [loading, setLoading] = useState(true);
  const [expandedSeries, setExpandedSeries] = useState<Set<number>>(new Set());
  
  // Modals state
  const [scrapingItem, setScrapingItem] = useState<{id: number, type: 'series' | 'book', title: string} | null>(null);
  const [viewingItem, setViewingItem] = useState<{id: number, type: 'series' | 'book'} | null>(null);
  const [autoScrapeSeriesId, setAutoScrapeSeriesId] = useState<number | null>(null);
  const [deletingSeriesId, setDeletingSeriesId] = useState<number | null>(null);
  
  // Sorting state
  const [sortField, setSortField] = useState<SortField>('title');
  const [sortOrder, setSortOrder] = useState<SortOrder>('asc');

  const cleanQuery = (raw: string) => {
    // Remove content inside [], (), {}, etc.
    let cleaned = raw.replace(/\[.*?\]|\(.*?\)|{.*?}/g, ' ');
    // Remove extensions
    cleaned = cleaned.replace(/\.[^/.]+$/, "");
    // Remove multiple spaces
    cleaned = cleaned.trim().replace(/\s+/g, ' ');
    return cleaned || raw;
  };

  const openScrapeModal = (item: any, type: 'series' | 'book', e?: React.MouseEvent) => {
    if (e) e.stopPropagation();
    let initialTitle = item.title;
    if (type === 'book') {
      initialTitle = cleanQuery(item.filename);
    }
    setScrapingItem({ id: item.ID, type, title: initialTitle });
  };

  const openDetails = (id: number, type: 'series' | 'book', e?: React.MouseEvent) => {
    if (e) e.stopPropagation();
    console.log(`[LibraryPage] Opening details for ${type} ${id}`);
    setViewingItem({ id, type });
  };

  const fetchLibrary = async () => {
    setLoading(true);
    try {
      const res = await axios.get('/api/manga');
      setSeriesList(res.data);
    } catch (err) {
      console.error(err);
    }
    setLoading(false);
  };

  const handleScan = async () => {
    try {
      await axios.post('/api/library/scan');
      showToast('Scan started in the background!');
    } catch (err) {
      console.error(err);
      showToast('Failed to start scan.', 'error');
    }
  };

  const handleClean = async () => {
    try {
      await axios.post('/api/library/clean');
      showToast('Database cleaning started!');
      setTimeout(fetchLibrary, 1000);
    } catch (err) {
      console.error(err);
      showToast('Failed to clean database.', 'error');
    }
  };

  const handleConfirmDelete = async () => {
    if (!deletingSeriesId) return;
    try {
      await axios.delete(`/api/manga/${deletingSeriesId}`);
      showToast('Series removed from database.');
      fetchLibrary();
    } catch (err) {
      console.error(err);
      showToast('Failed to remove series.', 'error');
    }
    setDeletingSeriesId(null);
  };

  const toggleExpand = (id: number) => {
    const newSet = new Set(expandedSeries);
    if (newSet.has(id)) newSet.delete(id);
    else newSet.add(id);
    setExpandedSeries(newSet);
  };

  const handleSort = (field: SortField) => {
    if (sortField === field) {
      setSortOrder(sortOrder === 'asc' ? 'desc' : 'asc');
    } else {
      setSortField(field);
      setSortOrder('asc');
    }
  };

  const sortedSeries = useMemo(() => {
    return [...seriesList].sort((a, b) => {
      let valA: any = a[sortField as keyof Series];
      let valB: any = b[sortField as keyof Series];
      
      if (sortField === 'bookCount') {
        valA = a.books?.length || 0;
        valB = b.books?.length || 0;
      }

      if (valA < valB) return sortOrder === 'asc' ? -1 : 1;
      if (valA > valB) return sortOrder === 'asc' ? 1 : -1;
      return 0;
    });
  }, [seriesList, sortField, sortOrder]);

  useEffect(() => {
    fetchLibrary();
  }, []);

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <div>
          <h2 className="text-3xl font-black text-gray-900 flex items-center">
            <LayoutList className="mr-3 text-blue-600 w-8 h-8" /> Manga Library
          </h2>
          <p className="text-sm text-gray-500 mt-1 font-medium">Manage your collection at series and book level</p>
        </div>
        <div className="flex gap-3">
          <button
            onClick={fetchLibrary}
            className="bg-white p-2.5 rounded-xl shadow-sm border border-gray-200 hover:bg-gray-50 transition-all active:scale-95"
            title="Refresh List"
          >
            <RefreshCw className={`w-5 h-5 text-gray-600 ${loading ? 'animate-spin' : ''}`} />
          </button>
          <button
            onClick={handleClean}
            className="bg-white text-red-500 border border-red-200 px-5 py-2.5 rounded-xl flex items-center hover:bg-red-50 transition-all shadow-sm font-bold active:scale-95"
          >
            <Eraser className="w-5 h-5 mr-2" />
            Clean Library
          </button>
          <button
            onClick={handleScan}
            className="bg-blue-600 text-white px-5 py-2.5 rounded-xl flex items-center hover:bg-blue-700 transition-all shadow-md shadow-blue-100 font-bold active:scale-95"
          >
            <FolderOpen className="w-5 h-5 mr-2" />
            Update Library
          </button>
        </div>
      </div>

      <div className="bg-white rounded-2xl shadow-sm border border-gray-200 overflow-hidden">
        <div className="overflow-x-auto">
          <table className="w-full text-left border-collapse">
            <thead>
              <tr className="bg-gray-50/80 border-b border-gray-200 select-none">
                <th className="px-6 py-4 w-10"></th>
                <th className="px-6 py-4 cursor-pointer hover:bg-gray-100 transition-colors" onClick={() => handleSort('title')}>
                  <div className="flex items-center text-xs font-bold text-gray-500 uppercase tracking-wider">
                    Title / Path <SortIcon field="title" sortField={sortField} sortOrder={sortOrder} />
                  </div>
                </th>
                <th className="px-6 py-4 cursor-pointer hover:bg-gray-100 transition-colors" onClick={() => handleSort('author')}>
                  <div className="flex items-center text-xs font-bold text-gray-500 uppercase tracking-wider">
                    Author <SortIcon field="author" sortField={sortField} sortOrder={sortOrder} />
                  </div>
                </th>
                <th className="px-6 py-4 cursor-pointer hover:bg-gray-100 transition-colors" onClick={() => handleSort('bookCount')}>
                  <div className="flex items-center text-xs font-bold text-gray-500 uppercase tracking-wider">
                    Books <SortIcon field="bookCount" sortField={sortField} sortOrder={sortOrder} />
                  </div>
                </th>
                <th className="px-6 py-4 cursor-pointer hover:bg-gray-100 transition-colors" onClick={() => handleSort('status')}>
                  <div className="flex items-center text-xs font-bold text-gray-500 uppercase tracking-wider">
                    Status <SortIcon field="status" sortField={sortField} sortOrder={sortOrder} />
                  </div>
                </th>
                <th className="px-6 py-4 text-right text-xs font-bold text-gray-500 uppercase tracking-wider w-80">Actions</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-100">
              {sortedSeries.map((series) => (
                <React.Fragment key={series.ID}>
                  <tr className="hover:bg-blue-50/30 transition-colors group">
                    <td className="px-6 py-5 text-center">
                      <button onClick={() => toggleExpand(series.ID)} className="text-gray-300 hover:text-blue-600 transition-colors">
                        {expandedSeries.has(series.ID) ? <ChevronDown className="w-5 h-5" /> : <ChevronRight className="w-5 h-5" />}
                      </button>
                    </td>
                    <td className="px-6 py-5 cursor-pointer" onClick={(e) => openDetails(series.ID, 'series', e)}>
                      <div className="font-bold text-gray-900 group-hover:text-blue-700 transition-colors">{series.title}</div>
                      <div className="text-[10px] text-gray-400 font-mono truncate max-w-[200px] mt-0.5">{series.path}</div>
                    </td>
                    <td className="px-6 py-5 text-sm text-gray-600 font-medium">{series.author || <span className="text-gray-300 italic">Unknown</span>}</td>
                    <td className="px-6 py-5 text-sm text-gray-500 font-bold">{series.books?.length || 0}</td>
                    <td className="px-6 py-5">
                      <span className={`px-2.5 py-1 text-[10px] font-black uppercase tracking-wider rounded-lg shadow-sm ${
                        series.status === 'Scraped' ? 'bg-green-100 text-green-700 border border-green-200' : 'bg-yellow-50 text-yellow-700 border border-yellow-100'
                      }`}>
                        {series.status || 'Unscraped'}
                      </span>
                    </td>
                    <td className="px-6 py-5">
                      <div className="flex justify-end gap-2">
                        <button 
                          onClick={(e) => openDetails(series.ID, 'series', e)}
                          className="px-3 py-1.5 bg-white border border-gray-200 text-gray-600 hover:bg-gray-50 rounded-lg text-[10px] font-black transition-all shadow-sm flex items-center"
                          title="View Details"
                        >
                          <Eye className="w-3.5 h-3.5 mr-1.5" /> DETAILS
                        </button>
                        <button 
                          onClick={(e) => { e.stopPropagation(); setAutoScrapeSeriesId(series.ID); }}
                          className="px-3 py-1.5 bg-white border border-purple-200 text-purple-600 hover:bg-purple-50 rounded-lg text-[10px] font-black transition-all shadow-sm flex items-center"
                          title="Auto Scrape All Books"
                        >
                          <Wand2 className="w-3.5 h-3.5 mr-1.5" /> AUTO
                        </button>
                        <button 
                          onClick={(e) => openScrapeModal(series, 'series', e)}
                          className="px-3 py-1.5 bg-white border border-blue-200 text-blue-600 hover:bg-blue-50 rounded-lg text-[10px] font-black transition-all shadow-sm flex items-center"
                        >
                          <Search className="w-3.5 h-3.5 mr-1.5" /> SCRAPE
                        </button>
                        <button 
                          onClick={(e) => { e.stopPropagation(); setDeletingSeriesId(series.ID); }}
                          className="px-3 py-1.5 bg-white border border-red-200 text-red-500 hover:bg-red-50 rounded-lg text-[10px] font-black transition-all shadow-sm flex items-center"
                          title="Remove from Database"
                        >
                          <Trash2 className="w-3.5 h-3.5 mr-1.5" /> REMOVE
                        </button>
                      </div>
                    </td>
                  </tr>
                  {expandedSeries.has(series.ID) && (
                    <tr>
                      <td colSpan={6} className="bg-gray-50/50 px-16 py-6 border-l-4 border-blue-500">
                        <div className="space-y-3">
                          <h4 className="text-[10px] font-black text-gray-400 uppercase tracking-[0.2em] mb-4 flex items-center">
                            <Book className="w-3 h-3 mr-2 text-blue-400" /> Files in series
                          </h4>
                          {series.books.map((book) => (
                            <div key={book.ID} className="flex items-center justify-between p-3 bg-white rounded-xl border border-gray-100 shadow-sm hover:border-blue-200 transition-all group/item">
                              <div className="flex items-center flex-1 min-w-0 cursor-pointer" onClick={(e) => openDetails(book.ID, 'book', e)}>
                                <div className="w-8 h-8 bg-blue-50 rounded-lg flex items-center justify-center mr-3 flex-shrink-0">
                                  <Book className="w-4 h-4 text-blue-400" />
                                </div>
                                <div className="truncate">
                                  <div className="text-sm font-bold text-gray-700 group-hover/item:text-blue-600 transition-colors">{book.filename}</div>
                                  <div className="text-[10px] text-gray-400 mt-0.5 font-medium">{book.author || 'Unknown Author'} • Vol. {book.volume || '?'}</div>
                                </div>
                              </div>
                              <div className="flex items-center gap-3 ml-4">
                                <span className={`text-[10px] font-bold px-2 py-0.5 rounded-md ${book.status === 'Scraped' ? 'text-green-600 bg-green-50' : 'text-gray-400 bg-gray-50'}`}>
                                  {book.status || 'Pending'}
                                </span>
                                <div className="flex gap-1.5">
                                  <button 
                                    onClick={(e) => openDetails(book.ID, 'book', e)}
                                    className="px-2.5 py-1.5 bg-white border border-gray-200 text-gray-600 hover:bg-gray-100 rounded-lg text-[10px] font-black transition-all shadow-sm flex items-center"
                                  >
                                    <Eye className="w-3 h-3 mr-1.5" /> DETAILS
                                  </button>
                                  <button 
                                    onClick={(e) => openScrapeModal(book, 'book', e)}
                                    className="px-2.5 py-1.5 bg-white border border-blue-100 text-blue-600 hover:bg-blue-600 hover:text-white rounded-lg text-[10px] font-black transition-all shadow-sm flex items-center"
                                  >
                                    <Search className="w-3 h-3 mr-1.5" /> SCRAPE
                                  </button>
                                </div>
                              </div>
                            </div>
                          ))}
                        </div>
                      </td>
                    </tr>
                  )}
                </React.Fragment>
              ))}
            </tbody>
          </table>
          {seriesList.length === 0 && !loading && (
            <div className="text-center py-32 bg-gray-50/50">
              <div className="w-20 h-20 bg-white rounded-3xl shadow-sm flex items-center justify-center mx-auto mb-6">
                <FolderOpen className="w-10 h-10 text-gray-200" />
              </div>
              <p className="text-gray-400 font-bold">Your library is empty</p>
              <p className="text-sm text-gray-300 mt-1">Add a folder in Settings to get started</p>
            </div>
          )}
        </div>
      </div>

      {scrapingItem && (
        <ScrapeModal
          itemId={scrapingItem.id}
          itemType={scrapingItem.type}
          initialTitle={scrapingItem.title}
          onClose={() => setScrapingItem(null)}
          onScraped={fetchLibrary}
        />
      )}

      {viewingItem && (
        <DetailModal
          key={`${viewingItem.type}-${viewingItem.id}`}
          itemId={viewingItem.id}
          itemType={viewingItem.type}
          onClose={() => setViewingItem(null)}
          onSaved={() => {
            setViewingItem(null);
            fetchLibrary();
          }}
        />
      )}

      {autoScrapeSeriesId && (
        <AutoScrapeModal
          seriesId={autoScrapeSeriesId}
          onClose={() => setAutoScrapeSeriesId(null)}
          onComplete={fetchLibrary}
        />
      )}

      {deletingSeriesId && (
        <ConfirmModal
          title="Remove from Database?"
          message="Are you sure you want to remove this series from the database? Local files will NOT be deleted. You can re-scan later."
          confirmText="Yes, Remove"
          isDanger={true}
          onConfirm={handleConfirmDelete}
          onCancel={() => setDeletingSeriesId(null)}
        />
      )}
    </div>
  );
};

export default LibraryPage;
