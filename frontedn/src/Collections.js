import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';

function Collections({ token, onLogout }) {
  const [collections, setCollections] = useState([]);
  const [error, setError] = useState(null);
  const [newCollectionName, setNewCollectionName] = useState('');
  const navigate = useNavigate();

  useEffect(() => {
    fetchCollections();
  }, []);

  const fetchCollections = async () => {
    setError(null);
    try {
      const response = await fetch('/api/collections', {
        headers: { Authorization: `Bearer ${token}` },
      });
      if (!response.ok) {
        if (response.status === 401) {
          onLogout();
        }
        const data = await response.json();
        setError(data.error || 'Failed to fetch collections');
        return;
      }
      const data = await response.json();
      setCollections(data);
    } catch (err) {
      setError('Network error');
    }
  };

  const handleCollectionClick = (collectionId) => {
    navigate(`/collections/${collectionId}`);
  };

  const handleCreateCollection = async () => {
    if (!newCollectionName.trim()) {
      setError('Collection name cannot be empty');
      return;
    }
    setError(null);
    try {
      const response = await fetch('/api/collections', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify({ name: newCollectionName }),
      });
      if (!response.ok) {
        if (response.status === 401) {
          onLogout();
        }
        const data = await response.json();
        setError(data.error || 'Failed to create collection');
        return;
      }
      setNewCollectionName('');
      fetchCollections();
    } catch (err) {
      setError('Network error');
    }
  };

  return (
    <div style={{ maxWidth: 600, margin: 'auto', padding: 20 }}>
      <h2>Your Collections</h2>
      {error && <div style={{ color: 'red' }}>{error}</div>}
      <div style={{ marginBottom: 20 }}>
        <input
          type="text"
          placeholder="New collection name"
          value={newCollectionName}
          onChange={(e) => setNewCollectionName(e.target.value)}
          style={{ padding: 8, width: '70%', marginRight: 10 }}
        />
        <button onClick={handleCreateCollection}>Create</button>
      </div>
      <ul style={{ listStyle: 'none', padding: 0 }}>
        {collections.map((col) => (
          <li
            key={col.id}
            onClick={() => handleCollectionClick(col.id)}
            style={{
              padding: '10px 15px',
              cursor: 'pointer',
              backgroundColor: '#f0f0f0',
              borderRadius: 6,
              marginBottom: 8,
              border: '1px solid #ccc',
              transition: 'background-color 0.2s',
            }}
            onMouseEnter={e => e.currentTarget.style.backgroundColor = '#ddd'}
            onMouseLeave={e => e.currentTarget.style.backgroundColor = '#f0f0f0'}
          >
            {col.name}
          </li>
        ))}
      </ul>
      <button onClick={onLogout} style={{ marginTop: 20 }}>Logout</button>
    </div>
  );
}

export default Collections;
