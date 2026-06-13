import React, { useEffect, useState, useRef, useCallback } from 'react';
import { RefreshCw, FolderOpen, ArrowUp, ArrowDown, LayoutList, Eraser, Book, Loader2 } from 'lucide-react';
import { api } from '../api';
import type { MangaSeries } from '../api';

import ScrapeModal from '../components/ScrapeModal';
import DetailModal from '../components/DetailModal';
import AutoScrapeModal from '../components/AutoScrapeModal';
import ConfirmModal from '../components/ConfirmModal';
import SeriesRow from '../components/library/SeriesRow';
import BookRow from '../components/library/BookRow';

import { showToast } from '../components/Toast';

type SortField = 'title' | 'author' | 'status' | 'bookCount';
type SortOrder = 'asc' | 'desc';

const SortIcon = ({ field, sortField, sortOrder }: { field: SortField, sortField: SortField, sortOrder: SortOrder }) => {
  if (sortField !== field) return null;
  return sortOrder === 'asc' ? <ArrowUp className="ml-1 w-4 h-4 text-blue-500" /> : <ArrowDown className="ml-1 w-4 h-4 text-blue-500" />;
};

const LibraryPage: React.FC = () => {
  const [seriesList, setSeriesList] = useState<MangaSeries[]>([]);
  const [loading, setLoading] = useState(false);
  const [page, setPage] = useState(1);
  const [hasMore, setHasMore] = useState(true);
  const [expandedSeries, setExpandedSeries] = useState<Set<number>>(new Set());
  
  // Modals state
  const [scrapingItem, setScrapingItem] = useState<{id: number, type: 'series' | 'book', title: string} | null>(null);
  const [viewingItem, setViewingItem] = useState<{id: number, type: 'series' | 'book'} | null>(null);
  const [autoScrapeSeriesId, setAutoScrapeSeriesId] = useState<number | null>(null);
  const [deletingSeriesId, setDeletingSeriesId] = useState<number | null>(null);
  
  // Sorting state
  const [sortField, setSortField] = useState<SortField>('title');
  const [sortOrder, setSortOrder] = useState<SortOrder>('asc');

  // Observer for infinite scroll
  const observer = useRef<IntersectionObserver | null>(null);
  const lastElementRef = useCallback((node: any) => {
    if (loading) return;
    if (observer.current) observer.current.disconnect();
    observer.current = new IntersectionObserver(entries => {
      if (entries[0].isIntersecting && hasMore) {
        setPage(prev => prev + 1);
      }
    });
    if (node) observer.current.observe(node);
  }, [loading, hasMore]);

  const openScrapeModal = (item: any, type: 'series' | 'book') => {
    const initialTitle = type === 'book' ? item.filename : item.title;
    setScrapingItem({ id: item.ID, type, title: initialTitle });
  };

  const fetchLibrary = async (pageNum: number, isNew: boolean = false) => {
    if (loading && !isNew) return;
    setLoading(true);
    try {
      // Pass sortField and sortOrder eventually if backend supports it. 
      // For now we ordered by title asc in backend.
      const res = await api.getSeriesList(pageNum, 20);
      const newData = res.data;
      
      if (isNew) {
        setSeriesList(newData);
      } else {
        setSeriesList(prev => [...prev, ...newData]);
      }
      
      setHasMore(newData.length === 20);
    } catch (err) {
      console.error(err);
    }
    setLoading(false);
  };

  const handleRefresh = () => {
    setPage(1);
    setHasMore(true);
    fetchLibrary(1, true);
  };

  const handleScan = async () => {
    try {
      await api.scanLibrary();
      showToast('Scan started in the background!');
    } catch (err) {
      console.error(err);
      showToast('Failed to start scan.', 'error');
    }
  };

  const handleClean = async () => {
    try {
      await api.cleanLibrary();
      showToast('Database cleaning started!');
      setTimeout(handleRefresh, 1000);
    } catch (err) {
      console.error(err);
      showToast('Failed to clean database.', 'error');
    }
  };

  const handleConfirmDelete = async () => {
    if (!deletingSeriesId) return;
    try {
      await api.deleteSeries(deletingSeriesId);
      showToast('Series removed from database.');
      handleRefresh();
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

  useEffect(() => {
    if (page > 1) {
      fetchLibrary(page);
    }
  }, [page]);

  useEffect(() => {
    handleRefresh();
  }, [sortField, sortOrder]);

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
            onClick={handleRefresh}
            className="bg-white p-2.5 rounded-xl shadow-sm border border-gray-200 hover:bg-gray-50 transition-all active:scale-95"
            title="Refresh List"
          >
            <RefreshCw className={`w-5 h-5 text-gray-600 ${loading && page === 1 ? 'animate-spin' : ''}`} />
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
              {seriesList.map((series, index) => (
                <React.Fragment key={series.ID}>
                  <SeriesRow 
                    series={series}
                    isExpanded={expandedSeries.has(series.ID)}
                    onToggleExpand={() => toggleExpand(series.ID)}
                    onViewDetails={() => setViewingItem({ id: series.ID, type: 'series' })}
                    onAutoScrape={() => setAutoScrapeSeriesId(series.ID)}
                    onScrape={() => openScrapeModal(series, 'series')}
                    onDelete={() => setDeletingSeriesId(series.ID)}
                    {...(index === seriesList.length - 1 ? { ref: lastElementRef } : {})}
                  />
                  {expandedSeries.has(series.ID) && (
                    <tr>
                      <td colSpan={6} className="bg-gray-50/50 px-16 py-6 border-l-4 border-blue-500">
                        <div className="space-y-3">
                          <h4 className="text-[10px] font-black text-gray-400 uppercase tracking-[0.2em] mb-4 flex items-center">
                            <Book className="w-3 h-3 mr-2 text-blue-400" /> Files in series
                          </h4>
                          {series.books.map((book) => (
                            <BookRow 
                              key={book.ID}
                              book={book}
                              onViewDetails={() => setViewingItem({ id: book.ID, type: 'book' })}
                              onScrape={() => openScrapeModal(book, 'book')}
                            />
                          ))}
                        </div>
                      </td>
                    </tr>
                  )}
                </React.Fragment>
              ))}
            </tbody>
          </table>
          {loading && (
            <div className="flex justify-center py-8">
              <Loader2 className="w-6 h-6 animate-spin text-blue-500" />
            </div>
          )}
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
          onScraped={handleRefresh}
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
            handleRefresh();
          }}
        />
      )}

      {autoScrapeSeriesId && (
        <AutoScrapeModal
          seriesId={autoScrapeSeriesId}
          onClose={() => setAutoScrapeSeriesId(null)}
          onComplete={handleRefresh}
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
