import { useState, useEffect } from 'react';

interface Tenant {
  id: string;
  name: string;
  slug: string;
  created_at: string;
}

const INITIAL_TENANTS: Tenant[] = [
  { id: 'tn_1', name: 'Acme Corporation', slug: 'acme', created_at: '2026-01-15T08:00:00Z' },
  { id: 'tn_2', name: 'Stark Industries', slug: 'stark', created_at: '2026-02-20T10:30:00Z' },
  { id: 'tn_3', name: 'Wayne Enterprises', slug: 'wayne', created_at: '2026-03-05T16:45:00Z' }
];

export default function Tenants() {
  const [tenants, setTenants] = useState<Tenant[]>(INITIAL_TENANTS);
  const [loading, setLoading] = useState(true);
  const [searchTerm, setSearchTerm] = useState('');
  const [showAddModal, setShowAddModal] = useState(false);
  const [newName, setNewName] = useState('');
  const [newSlug, setNewSlug] = useState('');

  useEffect(() => {
    const timer = setTimeout(() => setLoading(false), 400);
    return () => clearTimeout(timer);
  }, []);

  const handleAddTenant = (e: React.FormEvent) => {
    e.preventDefault();
    if (!newName) return;
    const slug = newSlug || newName.toLowerCase().replace(/[^a-z0-9]+/g, '-').replace(/(^-|-$)/g, '');
    const newTenant: Tenant = {
      id: `tn_${Date.now()}`,
      name: newName,
      slug: slug,
      created_at: new Date().toISOString()
    };
    setTenants([newTenant, ...tenants]);
    setNewName('');
    setNewSlug('');
    setShowAddModal(false);
  };

  const handleDeleteTenant = (id: string) => {
    if (window.confirm('Are you sure you want to delete this tenant?')) {
      setTenants(tenants.filter((t) => t.id !== id));
    }
  };

  const filteredTenants = tenants.filter(
    (tenant) =>
      tenant.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
      tenant.slug.toLowerCase().includes(searchTerm.toLowerCase())
  );

  return (
    <div className="animate-fade-in">
      <div className="sm:flex sm:items-center sm:justify-between mb-8">
        <div>
          <h1 className="text-3xl font-bold text-gray-900 tracking-tight">Tenants</h1>
          <p className="mt-2 text-gray-600">
            Manage organizations and their members.
          </p>
        </div>
        <div className="mt-4 sm:mt-0">
          <button
            type="button"
            onClick={() => setShowAddModal(true)}
            className="inline-flex items-center px-4 py-2.5 border border-transparent rounded-xl text-sm font-semibold text-white bg-primary-600 hover:bg-primary-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-primary-500 transition-all duration-200 shadow-sm cursor-pointer"
          >
            <svg className="w-5 h-5 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2.5} d="M12 4v16m8-8H4" />
            </svg>
            Create Tenant
          </button>
        </div>
      </div>

      <div className="mb-6">
        <div className="relative">
          <div className="absolute inset-y-0 left-0 pl-3.5 flex items-center pointer-events-none">
            <svg className="h-5 w-5 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
            </svg>
          </div>
          <input
            type="text"
            placeholder="Search tenants by name or slug..."
            className="block w-full pl-11 pr-4 py-3 border border-gray-200 rounded-xl focus:ring-2 focus:ring-primary-500 focus:border-transparent focus:outline-none transition-all duration-200 bg-white shadow-xs"
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
          />
        </div>
      </div>

      <div className="bg-white rounded-2xl shadow-sm border border-gray-100 overflow-hidden">
        <table className="min-w-full divide-y divide-gray-100">
          <thead className="bg-gray-50/50">
            <tr>
              <th scope="col" className="px-6 py-4 text-left text-xs font-semibold text-gray-500 uppercase tracking-wider">
                Tenant Name
              </th>
              <th scope="col" className="px-6 py-4 text-left text-xs font-semibold text-gray-500 uppercase tracking-wider">
                Slug identifier
              </th>
              <th scope="col" className="px-6 py-4 text-left text-xs font-semibold text-gray-500 uppercase tracking-wider">
                Created At
              </th>
              <th scope="col" className="px-6 py-4 text-right text-xs font-semibold text-gray-500 uppercase tracking-wider">
                Actions
              </th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-100 bg-white">
            {loading ? (
              <tr>
                <td colSpan={4} className="px-6 py-16 text-center text-sm text-gray-500">
                  <div className="flex items-center justify-center space-x-3">
                    <svg className="animate-spin h-5 w-5 text-primary-600" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                      <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                      <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                    </svg>
                    <span className="font-medium text-gray-600">Loading tenants...</span>
                  </div>
                </td>
              </tr>
            ) : filteredTenants.length === 0 ? (
              <tr>
                <td colSpan={4} className="px-6 py-16 text-center text-sm text-gray-400">
                  <svg className="mx-auto h-12 w-12 text-gray-300 mb-3" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4" />
                  </svg>
                  <p className="font-medium">No tenants found</p>
                </td>
              </tr>
            ) : (
              filteredTenants.map((tenant) => (
                <tr key={tenant.id} className="hover:bg-gray-50/55 transition-colors duration-150">
                  <td className="px-6 py-4.5 whitespace-nowrap">
                    <div className="flex items-center">
                      <div className="flex-shrink-0 h-10 w-10">
                        <div className="h-10 w-10 rounded-xl bg-green-50 border border-green-100 flex items-center justify-center">
                          <svg className="w-5 h-5 text-green-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4" />
                          </svg>
                        </div>
                      </div>
                      <div className="ml-4">
                        <div className="text-sm font-semibold text-gray-900">{tenant.name}</div>
                      </div>
                    </div>
                  </td>
                  <td className="px-6 py-4.5 whitespace-nowrap">
                    <code className="px-2.5 py-1.5 bg-gray-100/70 border border-gray-200/50 rounded-lg text-xs font-semibold text-gray-700 tracking-wide font-mono">
                      {tenant.slug}
                    </code>
                  </td>
                  <td className="px-6 py-4.5 whitespace-nowrap text-sm text-gray-500 font-medium">
                    {new Date(tenant.created_at).toLocaleDateString()}
                  </td>
                  <td className="px-6 py-4.5 whitespace-nowrap text-right text-sm font-medium">
                    <button 
                      onClick={() => handleDeleteTenant(tenant.id)}
                      className="text-xs font-semibold px-3 py-1.5 rounded-lg border border-red-200 text-red-600 hover:bg-red-50 transition-all duration-200 hover:shadow-xs cursor-pointer"
                    >
                      Delete
                    </button>
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>

      {showAddModal && (
        <div className="fixed inset-0 bg-gray-900/40 backdrop-blur-xs flex items-center justify-center z-50 animate-fade-in">
          <div className="bg-white rounded-2xl shadow-xl max-w-md w-full mx-4 border border-gray-100 overflow-hidden">
            <div className="px-6 py-5 border-b border-gray-50 flex items-center justify-between">
              <h3 className="text-lg font-bold text-gray-900">Create New Tenant</h3>
              <button 
                onClick={() => setShowAddModal(false)}
                className="text-gray-400 hover:text-gray-600 rounded-lg p-1 hover:bg-gray-55"
              >
                <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                </svg>
              </button>
            </div>
            <form onSubmit={handleAddTenant}>
              <div className="p-6 space-y-4">
                <div>
                  <label className="block text-sm font-semibold text-gray-700 mb-1.5">Tenant Name</label>
                  <input
                    type="text"
                    required
                    className="block w-full px-3.5 py-2.5 border border-gray-200 rounded-xl focus:ring-2 focus:ring-primary-500 focus:border-transparent focus:outline-none transition-all duration-200 text-sm"
                    placeholder="Acme Co"
                    value={newName}
                    onChange={(e) => setNewName(e.target.value)}
                  />
                </div>
                <div>
                  <label className="block text-sm font-semibold text-gray-700 mb-1.5">Slug (Optional)</label>
                  <input
                    type="text"
                    className="block w-full px-3.5 py-2.5 border border-gray-200 rounded-xl focus:ring-2 focus:ring-primary-500 focus:border-transparent focus:outline-none transition-all duration-200 text-sm"
                    placeholder="acme-co"
                    value={newSlug}
                    onChange={(e) => setNewSlug(e.target.value)}
                  />
                </div>
              </div>
              <div className="px-6 py-4.5 bg-gray-50 border-t border-gray-100 flex justify-end space-x-3">
                <button
                  type="button"
                  onClick={() => setShowAddModal(false)}
                  className="px-4 py-2 border border-gray-200 rounded-xl text-sm font-semibold text-gray-600 bg-white hover:bg-gray-50 transition-all duration-200 cursor-pointer"
                >
                  Cancel
                </button>
                <button 
                  type="submit"
                  className="px-4 py-2 bg-primary-600 text-white rounded-xl text-sm font-semibold hover:bg-primary-700 transition-all duration-200 shadow-sm cursor-pointer"
                >
                  Create
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
