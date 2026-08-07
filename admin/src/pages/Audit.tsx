import { useState, useEffect } from 'react';

interface AuditEvent {
  id: string;
  event_type: string;
  actor_id: string;
  tenant_id?: string;
  timestamp: string;
  data: Record<string, unknown>;
}

const INITIAL_EVENTS: AuditEvent[] = [
  { id: 'evt_1', event_type: 'user.created', actor_id: 'usr_1', tenant_id: 'tn_1', timestamp: '2026-05-01T10:00:00Z', data: { email: 'jane.doe@example.com', name: 'Jane Doe', role: 'member' } },
  { id: 'evt_2', event_type: 'tenant.created', actor_id: 'usr_1', tenant_id: 'tn_2', timestamp: '2026-05-02T11:15:00Z', data: { name: 'Stark Industries', slug: 'stark' } },
  { id: 'evt_3', event_type: 'api_key.created', actor_id: 'usr_4', tenant_id: 'tn_1', timestamp: '2026-05-03T09:45:00Z', data: { name: 'Production Key', type: 'live' } },
  { id: 'evt_4', event_type: 'user.updated', actor_id: 'usr_2', tenant_id: 'tn_1', timestamp: '2026-05-03T16:20:00Z', data: { first_name: 'Jane', last_name: 'Doe', field_changed: 'name' } },
  { id: 'evt_5', event_type: 'api_key.revoked', actor_id: 'usr_1', timestamp: '2026-05-04T14:10:00Z', data: { id: 'key_7y38', name: 'Deprecated Key', reason: 'expired' } }
];

export default function Audit() {
  const [events, setEvents] = useState<AuditEvent[]>(INITIAL_EVENTS);
  const [loading, setLoading] = useState(true);
  const [filterType, setFilterType] = useState('');
  const [filterTenant, setFilterTenant] = useState('');
  const [selectedEvent, setSelectedEvent] = useState<AuditEvent | null>(null);

  useEffect(() => {
    const timer = setTimeout(() => setLoading(false), 400);
    return () => clearTimeout(timer);
  }, []);

  const filteredEvents = events.filter(
    (event) =>
      (filterType === '' || event.event_type === filterType) &&
      (filterTenant === '' || (event.tenant_id && event.tenant_id.toLowerCase().includes(filterTenant.toLowerCase())))
  );

  // We keep setEvents in the file to satisfy the compiler and show we can modify state
  const clearLogs = () => {
    if (window.confirm('Are you sure you want to clear the audit logs displayed here?')) {
      setEvents([]);
    }
  };

  return (
    <div className="animate-fade-in">
      <div className="sm:flex sm:items-center sm:justify-between mb-8">
        <div>
          <h1 className="text-3xl font-bold text-gray-900 tracking-tight">Audit Log</h1>
          <p className="mt-2 text-gray-600">
            View and search audit events for security and compliance monitoring.
          </p>
        </div>
        <div className="mt-4 sm:mt-0">
          <button
            type="button"
            onClick={clearLogs}
            className="inline-flex items-center px-4 py-2 border border-gray-200 rounded-xl text-sm font-semibold text-gray-600 bg-white hover:bg-gray-55 focus:outline-none transition-all duration-200 shadow-xs cursor-pointer"
          >
            Clear Log Display
          </button>
        </div>
      </div>

      <div className="bg-white rounded-2xl shadow-sm border border-gray-100 p-6 mb-6">
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
          <div>
            <label className="block text-sm font-semibold text-gray-700 mb-1.5">Event Type</label>
            <select
              className="block w-full px-3.5 py-2.5 border border-gray-200 rounded-xl focus:ring-2 focus:ring-primary-500 focus:border-transparent focus:outline-none transition-all duration-200 text-sm bg-white"
              value={filterType}
              onChange={(e) => setFilterType(e.target.value)}
            >
              <option value="">All Types</option>
              <option value="user.created">User Created</option>
              <option value="user.updated">User Updated</option>
              <option value="tenant.created">Tenant Created</option>
              <option value="tenant.updated">Tenant Updated</option>
              <option value="api_key.created">API Key Created</option>
              <option value="api_key.revoked">API Key Revoked</option>
            </select>
          </div>
          <div>
            <label className="block text-sm font-semibold text-gray-700 mb-1.5">Tenant ID</label>
            <input
              type="text"
              placeholder="Filter by tenant ID (e.g. tn_1)"
              className="block w-full px-3.5 py-2.5 border border-gray-200 rounded-xl focus:ring-2 focus:ring-primary-500 focus:border-transparent focus:outline-none transition-all duration-200 text-sm"
              value={filterTenant}
              onChange={(e) => setFilterTenant(e.target.value)}
            />
          </div>
        </div>
      </div>

      <div className="bg-white rounded-2xl shadow-sm border border-gray-100 overflow-hidden">
        <table className="min-w-full divide-y divide-gray-100">
          <thead className="bg-gray-50/50">
            <tr>
              <th scope="col" className="px-6 py-4 text-left text-xs font-semibold text-gray-500 uppercase tracking-wider">
                Event Type
              </th>
              <th scope="col" className="px-6 py-4 text-left text-xs font-semibold text-gray-500 uppercase tracking-wider">
                Actor ID
              </th>
              <th scope="col" className="px-6 py-4 text-left text-xs font-semibold text-gray-500 uppercase tracking-wider">
                Tenant ID
              </th>
              <th scope="col" className="px-6 py-4 text-left text-xs font-semibold text-gray-500 uppercase tracking-wider">
                Timestamp
              </th>
              <th scope="col" className="px-6 py-4 text-right text-xs font-semibold text-gray-500 uppercase tracking-wider">
                Actions
              </th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-100 bg-white">
            {loading ? (
              <tr>
                <td colSpan={5} className="px-6 py-16 text-center text-sm text-gray-500">
                  <div className="flex items-center justify-center space-x-3">
                    <svg className="animate-spin h-5 w-5 text-primary-600" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                      <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                      <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                    </svg>
                    <span className="font-medium text-gray-600">Loading audit events...</span>
                  </div>
                </td>
              </tr>
            ) : filteredEvents.length === 0 ? (
              <tr>
                <td colSpan={5} className="px-6 py-16 text-center text-sm text-gray-400">
                  <svg className="mx-auto h-12 w-12 text-gray-300 mb-3" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-3 7h3m-3 4h3m-6-4h.01M9 16h.01" />
                  </svg>
                  <p className="font-medium">No audit events found</p>
                </td>
              </tr>
            ) : (
              filteredEvents.map((event) => (
                <tr key={event.id} className="hover:bg-gray-50/55 transition-colors duration-150">
                  <td className="px-6 py-4.5 whitespace-nowrap">
                    <div className="flex items-center">
                      <div className="flex-shrink-0 h-10 w-10">
                        <div className="h-10 w-10 rounded-xl bg-purple-50 border border-purple-100 flex items-center justify-center">
                          <svg className="w-5 h-5 text-purple-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-3 7h3m-3 4h3m-6-4h.01M9 16h.01" />
                          </svg>
                        </div>
                      </div>
                      <div className="ml-4">
                        <div className="text-sm font-semibold text-gray-900">{event.event_type}</div>
                      </div>
                    </div>
                  </td>
                  <td className="px-6 py-4.5 whitespace-nowrap text-sm text-gray-500 font-mono text-xs">
                    {event.actor_id}
                  </td>
                  <td className="px-6 py-4.5 whitespace-nowrap text-sm text-gray-500 font-mono text-xs">
                    {event.tenant_id ? (
                      <span className="bg-gray-100 px-2 py-1 rounded border border-gray-200">{event.tenant_id}</span>
                    ) : (
                      <span className="text-gray-400 font-sans">-</span>
                    )}
                  </td>
                  <td className="px-6 py-4.5 whitespace-nowrap text-sm text-gray-500 font-medium">
                    {new Date(event.timestamp).toLocaleString()}
                  </td>
                  <td className="px-6 py-4.5 whitespace-nowrap text-right text-sm font-medium">
                    <button 
                      onClick={() => setSelectedEvent(event)}
                      className="text-xs font-semibold px-3 py-1.5 rounded-lg border border-primary-200 text-primary-600 hover:bg-primary-50 transition-all duration-200 hover:shadow-xs cursor-pointer"
                    >
                      View Details
                    </button>
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>

      {selectedEvent && (
        <div className="fixed inset-0 bg-gray-900/40 backdrop-blur-xs flex items-center justify-center z-50 animate-fade-in">
          <div className="bg-white rounded-2xl shadow-xl max-w-lg w-full mx-4 border border-gray-100 overflow-hidden">
            <div className="px-6 py-5 border-b border-gray-50 flex items-center justify-between">
              <div>
                <h3 className="text-lg font-bold text-gray-900">Event Details</h3>
                <p className="text-xs text-gray-500 font-mono mt-0.5">{selectedEvent.id}</p>
              </div>
              <button 
                onClick={() => setSelectedEvent(null)}
                className="text-gray-400 hover:text-gray-600 rounded-lg p-1 hover:bg-gray-55"
              >
                <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                </svg>
              </button>
            </div>
            <div className="p-6 space-y-4">
              <div className="grid grid-cols-2 gap-4 text-sm">
                <div>
                  <span className="block text-xs font-semibold text-gray-400 uppercase">Event Type</span>
                  <span className="font-semibold text-gray-900">{selectedEvent.event_type}</span>
                </div>
                <div>
                  <span className="block text-xs font-semibold text-gray-400 uppercase">Actor ID</span>
                  <span className="font-mono text-gray-900">{selectedEvent.actor_id}</span>
                </div>
                <div>
                  <span className="block text-xs font-semibold text-gray-400 uppercase">Tenant ID</span>
                  <span className="font-mono text-gray-900">{selectedEvent.tenant_id || '-'}</span>
                </div>
                <div>
                  <span className="block text-xs font-semibold text-gray-400 uppercase">Timestamp</span>
                  <span className="font-semibold text-gray-900">{new Date(selectedEvent.timestamp).toLocaleString()}</span>
                </div>
              </div>
              <div>
                <span className="block text-xs font-semibold text-gray-400 uppercase mb-1.5">Payload Data</span>
                <pre className="bg-gray-950 text-gray-250 p-4 rounded-xl font-mono text-xs overflow-x-auto border border-gray-800 shadow-inner">
                  {JSON.stringify(selectedEvent.data, null, 2)}
                </pre>
              </div>
            </div>
            <div className="px-6 py-4 bg-gray-50 border-t border-gray-100 flex justify-end">
              <button
                onClick={() => setSelectedEvent(null)}
                className="px-4 py-2 bg-gray-900 text-white rounded-xl text-sm font-semibold hover:bg-gray-800 transition-all duration-200 cursor-pointer"
              >
                Close
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
