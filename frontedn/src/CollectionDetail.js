import React, { useState, useEffect, useRef } from 'react';
import { useParams, useNavigate } from 'react-router-dom';

function StudySession({ cards, onExit, token, collectionId }) {
  const [index, setIndex] = useState(0);
  const [showBack, setShowBack] = useState(false);
  const current = cards[index];

  const handleResponse = async (difficulty) => {
    try {
      await fetch(`/api/collections/${collectionId}/cards/${current.id}/response`, {
        method: 'POST',
        headers: {
          Authorization: `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ difficulty }),
      });
    } catch (e) {
      console.error('Dummy response failed', e);
    }
    setShowBack(false);
    if (index + 1 < cards.length) setIndex(index + 1);
    else {
      alert('Study complete!');
      onExit();
    }
  };

  return (
    <div style={{
      position: 'fixed', top: 0, left: 0, width: '100vw', height: '100vh',
      backgroundColor: 'black', color: 'white', display: 'flex',
      flexDirection: 'column', alignItems: 'center', justifyContent: 'center',
      zIndex: 9999, padding: 20
    }}>
      <h2>Card {index + 1} of {cards.length}</h2>
      <div style={{ fontSize: 24, marginBottom: 20 }}>
        {showBack ? current.back : current.front}
      </div>
      {!showBack ? (
        <button onClick={() => setShowBack(true)}>Show Answer</button>
      ) : (
        <div>
          <button onClick={() => handleResponse('again')} style={{ marginRight: 10 }}>Again</button>
          <button onClick={() => handleResponse('hard')} style={{ marginRight: 10 }}>Hard</button>
          <button onClick={() => handleResponse('medium')} style={{ marginRight: 10 }}>Medium</button>
          <button onClick={() => handleResponse('easy')}>Easy</button>
        </div>
      )}
      <button onClick={onExit} style={{ marginTop: 30, fontSize: 12 }}>Exit</button>
    </div>
  );
}

function CollectionDetail({ token, onLogout }) {
  const { collectionId } = useParams();
  const navigate = useNavigate();
  const [cards, setCards] = useState([]);
  const [files, setFiles] = useState([]);
  const [collectionName, setCollectionName] = useState('');
  const [error, setError] = useState(null);
  const [file, setFile] = useState(null);
  const [uploading, setUploading] = useState(false);
  const [uploadError, setUploadError] = useState(null);
  const [uploadSuccess, setUploadSuccess] = useState(null);
  const [presignedUrl, setPresignedUrl] = useState(null);
  const [objectKey, setObjectKey] = useState(null);
  const [statusMessage, setStatusMessage] = useState('');
  const [processingFile, setProcessingFile] = useState(() => localStorage.getItem('processingFile') || null);
  const [studyMode, setStudyMode] = useState(false);
  const socketRef = useRef(null);

  useEffect(() => {
    fetchCollection();
    fetchFiles();
  }, [collectionId]);

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
      if (!response.ok) return;
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
      const res = await fetch(`/api/collections/${collectionId}/presigUrl`, {
        method: 'GET',
        headers: { Authorization: `Bearer ${token}` },
      });
      const data = await res.json();
      setPresignedUrl(data.url);
      setObjectKey(data.objectkey);
    } catch {
      setUploadError('Error getting presigned URL');
    }
  };

  const handleUpload = async () => {
    if (!file || !presignedUrl) {
      setUploadError('Missing file or URL');
      return;
    }
    setUploading(true);
    setUploadSuccess(null);
    setUploadError(null);
    setStatusMessage('Uploading...');
    setProcessingFile(file.name);
    localStorage.setItem('processingFile', file.name);
    localStorage.setItem('objectKey', objectKey);

    try {
      await fetch(presignedUrl, {
        method: 'PUT',
        headers: {
          'Content-Type': file.type,
          'Content-Length': file.size.toString(),
        },
        body: file,
      });

      const res = await fetch(`/api/collections/${collectionId}/verifyUpload`, {
        method: 'POST',
        headers: {
          Authorization: `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          file_name: file.name,
          object_key: objectKey,
          format: file.name.split('.').pop().toLowerCase(),
          error: '',
          status: 'success',
        }),
      });

      const socket = new WebSocket(`ws://localhost:8000/api/ws`);
      socketRef.current = socket;
      socket.onopen = () => socket.send(JSON.stringify({ fileID: objectKey }));
      socket.onmessage = (event) => {
        const msg = JSON.parse(event.data);
        if (msg.message) {
          setStatusMessage(msg.message);
          setUploadSuccess('Upload complete!');
          setProcessingFile(null);
          localStorage.removeItem('processingFile');
          localStorage.removeItem('objectKey');
          fetchCollection();
          fetchFiles();
          socket.close();
        }
      };
      socket.onerror = () => setUploadError('WebSocket failed');
    } catch (err) {
      setUploadError(err.message);
    } finally {
      setUploading(false);
    }
  };

  return (
    <div style={{ maxWidth: 800, margin: 'auto', padding: 20 }}>
      <h2>Collection: {collectionName}</h2>
      <button onClick={onLogout} style={{ marginBottom: 20, marginRight: 10 }}>Logout</button>
      <button onClick={() => navigate('/collections')} style={{ marginBottom: 20 }}>Back to Collections</button>
      {error && <div style={{ color: 'red' }}>{error}</div>}

      <h3>Files in this Collection</h3>
      {files.length === 0 ? <p>No files uploaded yet.</p> : (
        <ul style={{ listStyle: 'none', padding: 0 }}>
          {files.map((file) => (
            <li key={file.id} style={{ marginBottom: 8 }}>{file.filename}</li>
          ))}
        </ul>
      )}

      <h3>Upload File</h3>
      <input type="file" onChange={handleFileChange} />
      <button onClick={handleUpload} disabled={uploading || !file} style={{ marginLeft: 10 }}>
        {uploading ? 'Uploading...' : 'Upload'}
      </button>
      {uploadError && <div style={{ color: 'red', marginTop: 10 }}>{uploadError}</div>}
      {uploadSuccess && <div style={{ color: 'green', marginTop: 10 }}>{uploadSuccess}</div>}
      {statusMessage && <div style={{ marginTop: 10 }}>{statusMessage}</div>}

      <h3>Cards</h3>
      <p>{cards.length} card(s) in this collection.</p>
      {!studyMode && cards.length > 0 && (
        <button onClick={() => setStudyMode(true)}>Start Studying</button>
      )}
      {studyMode && (
        <StudySession
          cards={cards}
          onExit={() => setStudyMode(false)}
          token={token}
          collectionId={collectionId}
        />
      )}
    </div>
  );
}

export default CollectionDetail;
