import React, { useState, useEffect, useRef } from 'react';
import { useParams, useNavigate } from 'react-router-dom';

function CollectionDetail({ token, onLogout }) {
  const { collectionId } = useParams();
  const navigate = useNavigate();
  const [cards, setCards] = useState([]);
  const [files, setFiles] = useState([]); // NEW: Files state
  const [collectionName, setCollectionName] = useState('');
  const [error, setError] = useState(null);
  const [file, setFile] = useState(null);
  const [uploading, setUploading] = useState(false);
  const [uploadError, setUploadError] = useState(null);
  const [uploadSuccess, setUploadSuccess] = useState(null);
  const [presignedUrl, setPresignedUrl] = useState(null);
  const [objectKey, setObjectKey] = useState(null);
  const [statusMessage, setStatusMessage] = useState('');
  const [processingFile, setProcessingFile] = useState(() => {
    return localStorage.getItem('processingFile') || null;
  });

  const socketRef = useRef(null);

  useEffect(() => {
    fetchCollection();
  }, [collectionId]);

  useEffect(() => {
    fetchFiles(); // NEW: Fetch files separately
  }, [collectionId]);

  useEffect(() => {
    const savedFile = localStorage.getItem('processingFile');
    const savedObjectKey = localStorage.getItem('objectKey');
    if (savedFile && savedObjectKey && !socketRef.current) {
      const socket = new WebSocket(`ws://localhost:8000/api/ws`);
      socketRef.current = socket;

      socket.onopen = () => {
        socket.send(JSON.stringify({ fileID: savedObjectKey }));
      };
      socket.onmessage = (event) => {
        const msg = JSON.parse(event.data);
        if (msg.message) {
          setStatusMessage(msg.message);
          setUploadSuccess('File uploaded and cards created!');
          setProcessingFile(null);
          localStorage.removeItem('processingFile');
          localStorage.removeItem('objectKey');
          fetchCollection();
          fetchFiles(); // NEW: Refresh files when done
          socket.close();
          socketRef.current = null;
        }
      };
      socket.onerror = () => {
        setUploadError('WebSocket connection failed');
        setProcessingFile(null);
        localStorage.removeItem('processingFile');
        localStorage.removeItem('objectKey');
        socketRef.current = null;
      };
    }

    return () => {
      if (socketRef.current) {
        socketRef.current.close();
        socketRef.current = null;
      }
    };
  }, []);

  const fetchCollection = async () => {
    setError(null);
    try {
      const response = await fetch(`/api/collections/${collectionId}`, {
        headers: { Authorization: `Bearer ${token}` },
      });
      if (!response.ok) {
        if (response.status === 401) onLogout();
        const data = await response.json();
        setError(data.error || 'Failed to fetch collection');
        return;
      }
      const data = await response.json();
      setCollectionName(data.name || '');
      setCards(data.cards || []);
    } catch (err) {
      setError('Network error');
    }
  };

  const fetchFiles = async () => {
    try {
      const response = await fetch(`/api/collections/${collectionId}/files`, {
        headers: { Authorization: `Bearer ${token}` },
      });
      if (!response.ok) {
        return; // Optional: you can show error
      }
      const data = await response.json();
      setFiles(data || []);
    } catch (err) {
      console.error('Error fetching files:', err);
    }
  };

  const handleFileChange = async (e) => {
    const selectedFile = e.target.files[0];
    setFile(selectedFile);
    setUploadError(null);
    setUploadSuccess(null);
    setPresignedUrl(null);
    setObjectKey(null);

    if (!selectedFile) return;

    try {
      const presignResponse = await fetch(`/api/collections/${collectionId}/presigUrl`, {
        method: 'GET',
        headers: { Authorization: `Bearer ${token}` },
      });
      if (!presignResponse.ok) {
        const data = await presignResponse.json();
        setUploadError(data.error || 'Failed to get presigned URL');
        return;
      }
      const data = await presignResponse.json();
      setPresignedUrl(data.url);
      setObjectKey(data.objectkey);
    } catch (err) {
      setUploadError('Network error getting presigned URL');
    }
  };

  const handleUpload = async () => {
    if (!file || !presignedUrl) {
      setUploadError('Missing file or presigned URL.');
      return;
    }
    setUploading(true);
    setUploadError(null);
    setUploadSuccess(null);
    setStatusMessage('Creating cards...');
    setProcessingFile(file.name);
    localStorage.setItem('processingFile', file.name);
    localStorage.setItem('objectKey', objectKey);

    try {
      const putResponse = await fetch(presignedUrl, {
        method: 'PUT',
        headers: {
          'Content-Type': file.type,
          'Content-Length': file.size.toString(),
        },
        body: file,
      });
      if (!putResponse.ok) {
        setUploadError('Failed to upload file to storage');
        setUploading(false);
        setProcessingFile(null);
        localStorage.removeItem('processingFile');
        localStorage.removeItem('objectKey');
        return;
      }

      const verifyResponse = await fetch(`/api/collections/${collectionId}/verifyUpload`, {
        method: 'POST',
        headers: {
          Authorization: `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ file_name: file.name, object_key: objectKey,format:file.name.split('.').pop().toLowerCase(),   error: '', status: 'success' }),
      });
      if (!verifyResponse.ok) {
        let data;
        try {
          data = await verifyResponse.json();
        } catch {
          data = { error: 'Unknown error verifying upload' };
        }
        setUploadError(data.error || 'Failed to verify upload');
        setUploading(false);
        setProcessingFile(null);
        localStorage.removeItem('processingFile');
        localStorage.removeItem('objectKey');
        return;
      }

      const socket = new WebSocket(`ws://localhost:8000/api/ws`);
      socketRef.current = socket;
      socket.onopen = () => {
        socket.send(JSON.stringify({ fileID: objectKey }));
      };
      socket.onmessage = (event) => {
        const msg = JSON.parse(event.data);
        if (msg.message) {
          setStatusMessage(msg.message);
          setUploadSuccess('File uploaded and cards created!');
          setProcessingFile(null);
          localStorage.removeItem('processingFile');
          localStorage.removeItem('objectKey');
          fetchCollection();
          fetchFiles(); // NEW
          socket.close();
          socketRef.current = null;
        }
      };
      socket.onerror = () => {
        setUploadError('WebSocket connection failed');
        setProcessingFile(null);
        localStorage.removeItem('processingFile');
        localStorage.removeItem('objectKey');
        socketRef.current = null;
      };

      setFile(null);
      setPresignedUrl(null);
      setObjectKey(null);
    } catch (err) {
      setUploadError(`Network error during upload: ${err.message}`);
      setProcessingFile(null);
      localStorage.removeItem('processingFile');
      localStorage.removeItem('objectKey');
    } finally {
      setUploading(false);
    }
  };

  return (
    <div style={{ maxWidth: 800, margin: 'auto', padding: 20 }}>
      {processingFile && (
        <div style={{ position: 'fixed', top: 10, right: 10, padding: '10px 20px', backgroundColor: '#ffd700', borderRadius: 5 }}>
          Processing "{processingFile}"...
        </div>
      )}
      <h2>Collection: {collectionName}</h2>
      <button onClick={onLogout} style={{ marginBottom: 20, marginRight: 10 }}>Logout</button>
      <button onClick={() => navigate('/collections')} style={{ marginBottom: 20 }}>Back to Collections</button>
      {error && <div style={{ color: 'red' }}>{error}</div>}

      <h3>Files in this Collection</h3>
      {files.length === 0 && <p>No files uploaded yet.</p>}
      <ul style={{ listStyle: 'none', padding: 0 }}>
        {files.map((file) => (
          <li key={file.id} style={{ marginBottom: 8, padding: 8, borderBottom: '1px solid #ccc' }}>
            {file.filename}
          </li>
        ))}
      </ul>

      <h3>Cards</h3>
      {cards.length === 0 && <p>No cards in this collection.</p>}
      <ul style={{ listStyle: 'none', padding: 0 }}>
        {cards.map((card) => (
          <CardItem
            key={card.id}
            card={card}
            token={token}
            collectionId={collectionId}
            fetchCollection={fetchCollection}
            onLogout={onLogout}
          />
        ))}
      </ul>

      <h3>Upload File</h3>
      <input type="file" onChange={handleFileChange} />
      <button onClick={handleUpload} disabled={uploading || !file} style={{ marginLeft: 10 }}>
        {uploading ? 'Uploading...' : 'Upload'}
      </button>
      {uploadError && <div style={{ color: 'red', marginTop: 10 }}>{uploadError}</div>}
      {uploadSuccess && <div style={{ color: 'green', marginTop: 10 }}>{uploadSuccess}</div>}
      {statusMessage && <div style={{ marginTop: 10 }}>{statusMessage}</div>}
    </div>
  );
}


function CardItem({ card, token, collectionId, fetchCollection, onLogout }) {
  const [editing, setEditing] = useState(false);
  const [front, setFront] = useState(card.front);
  const [back, setBack] = useState(card.back);
  const [error, setError] = useState(null);

  const handleDeleteCard = async () => {
    if (!window.confirm('Are you sure you want to delete this card?')) return;
    try {
      const response = await fetch(`/api/collections/${collectionId}/cards/${card.id}`, {
        method: 'DELETE',
        headers: { Authorization: `Bearer ${token}` },
      });
      if (!response.ok) {
        if (response.status === 401) onLogout();
        const data = await response.json();
        setError(data.error || 'Failed to delete card');
        return;
      }
      fetchCollection();
    } catch (err) {
      setError('Network error');
    }
  };

  const handleUpdateCard = async () => {
    try {
      const response = await fetch(`/api/collections/${collectionId}/cards/${card.id}`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify({ front, back }),
      });
      if (!response.ok) {
        if (response.status === 401) onLogout();
        const data = await response.json();
        setError(data.error || 'Failed to update card');
        return;
      }
      setEditing(false);
      fetchCollection();
    } catch (err) {
      setError('Network error');
    }
  };

  return (
    <li style={{ marginBottom: 10, padding: 10, border: '1px solid #ccc', borderRadius: 4 }}>
      {editing ? (
        <div>
          <input
            type="text"
            value={front}
            onChange={(e) => setFront(e.target.value)}
            placeholder="Front"
            style={{ width: '100%', marginBottom: 8 }}
          />
          <input
            type="text"
            value={back}
            onChange={(e) => setBack(e.target.value)}
            placeholder="Back"
            style={{ width: '100%', marginBottom: 8 }}
          />
          <button onClick={handleUpdateCard} style={{ marginRight: 8 }}>Save</button>
          <button onClick={() => setEditing(false)}>Cancel</button>
        </div>
      ) : (
        <div>
          <div><strong>Front:</strong> {card.front}</div>
          <div><strong>Back:</strong> {card.back}</div>
          <button onClick={() => setEditing(true)} style={{ marginRight: 8 }}>Edit</button>
          <button onClick={handleDeleteCard}>Delete</button>
          {error && <div style={{ color: 'red' }}>{error}</div>}
        </div>
      )}
    </li>
  );
}
export default CollectionDetail;

