import React, { useState, useEffect } from 'react';

function Collections({ token, onLogout }) {
  const [collections, setCollections] = useState([]);
  const [selectedCollection, setSelectedCollection] = useState(null);
  const [cards, setCards] = useState([]);
  const [error, setError] = useState(null);

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
      if (data.length > 0) {
        selectCollection(data[0].id);
      }
    } catch (err) {
      setError('Network error');
    }
  };

  const selectCollection = async (collectionId) => {
    setSelectedCollection(collectionId);
    setError(null);
    try {
      const response = await fetch(`/api/collections/${collectionId}`, {
        headers: { Authorization: `Bearer ${token}` },
      });
      if (!response.ok) {
        if (response.status === 401) {
          onLogout();
        }
        const data = await response.json();
        setError(data.error || 'Failed to fetch cards');
        return;
      }
      const data = await response.json();
      setCards(data.cards || []);
    } catch (err) {
      setError('Network error');
    }
  };

  return (
    <div style={{ display: 'flex', maxWidth: 900, margin: 'auto', padding: 20 }}>
      <div style={{ width: 250, marginRight: 20 }}>
        <h3>Your Collections</h3>
        {error && <div style={{ color: 'red' }}>{error}</div>}
        <ul style={{ listStyle: 'none', padding: 0 }}>
          {collections.map((col) => (
            <li
              key={col.id}
              onClick={() => selectCollection(col.id)}
              style={{
                padding: '8px 12px',
                cursor: 'pointer',
                backgroundColor: selectedCollection === col.id ? '#ddd' : 'transparent',
                borderRadius: 4,
                marginBottom: 4,
              }}
            >
              {col.name}
            </li>
          ))}
        </ul>
        <button onClick={onLogout} style={{ marginTop: 20 }}>Logout</button>
      </div>
      <div style={{ flex: 1 }}>
        <h3>Cards</h3>
        {cards.length === 0 && <p>No cards in this collection.</p>}
        <ul style={{ listStyle: 'none', padding: 0 }}>
          {cards.map((card) => (
            <li key={card.id} style={{ marginBottom: 10, padding: 10, border: '1px solid #ccc', borderRadius: 4 }}>
              <div><strong>Front:</strong> {card.front}</div>
              <div><strong>Back:</strong> {card.back}</div>
            </li>
          ))}
        </ul>
      </div>
    </div>
  );
}

export default Collections;
