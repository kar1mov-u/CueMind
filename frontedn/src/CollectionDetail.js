import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';

function CollectionDetail({ token, onLogout }) {
  const { collectionId } = useParams();
  const navigate = useNavigate();
  const [cards, setCards] = useState([]);
  const [collectionName, setCollectionName] = useState('');
  const [error, setError] = useState(null);
  const [file, setFile] = useState(null);
  const [uploading, setUploading] = useState(false);
  const [uploadError, setUploadError] = useState(null);
  const [uploadSuccess, setUploadSuccess] = useState(null);
  const [presignedUrl, setPresignedUrl] = useState(null);
  const [objectKey, setObjectKey] = useState(null);

  useEffect(() => {
    fetchCollection();
  }, [collectionId]);

  const fetchCollection = async () => {
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

  const handleFileChange = async (e) => {
    const selectedFile = e.target.files[0];
    setFile(selectedFile);
    setUploadError(null);
    setUploadSuccess(null);
    setPresignedUrl(null);
    setObjectKey(null);

    if (!selectedFile) {
      return;
    }

    try {
      // Request presigned URL immediately on file selection
      const presignResponse = await fetch(`/api/collections/${collectionId}/presigUrl`, {
        method: 'GET',
        headers: {
          Authorization: `Bearer ${token}`,
        },
      });

      if (!presignResponse.ok) {
        const data = await presignResponse.json();
        setUploadError(data.error || 'Failed to get presigned URL');
        return;
      }

      const data = await presignResponse.json();
      setPresignedUrl(data.url);
      setObjectKey(data.objectKey);
    } catch (err) {
      setUploadError('Network error getting presigned URL');
    }
  };

  const handleUpload = async () => {
    if (!file) {
      setUploadError('Please select a file to upload.');
      return;
    }
    if (!presignedUrl) {
      setUploadError('No presigned URL available.');
      return;
    }
    setUploading(true);
    setUploadError(null);
    setUploadSuccess(null);

    try {
      // Upload file to S3 using stored presigned URL
      const putResponse = await fetch(presignedUrl, {
        method: 'PUT',
        headers: {
          'Content-Type': file.type,
        },
        body: file,
      });

      if (!putResponse.ok) {
        setUploadError('Failed to upload file to storage');
        setUploading(false);
        return;
      }

      // Notify backend of successful upload (verifyUpload)
      const verifyResponse = await fetch(`/api/collections/${collectionId}/verifyUpload`, {
        method: 'POST',
        headers: {
          Authorization: `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ fileName: file.name, objectKey }),
      });

      if (!verifyResponse.ok) {
        const data = await verifyResponse.json();
        setUploadError(data.error || 'Failed to verify upload');
        setUploading(false);
        return;
      }

      setUploadSuccess('File uploaded successfully');
      setFile(null);
      setPresignedUrl(null);
      setObjectKey(null);
      fetchCollection(); // Refresh cards/files if needed
    } catch (err) {
      setUploadError('Network error during upload');
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

      <h3>Upload File</h3>
      <input type="file" onChange={handleFileChange} />
      <button onClick={handleUpload} disabled={uploading || !file} style={{ marginLeft: 10 }}>
        {uploading ? 'Uploading...' : 'Upload'}
      </button>
      {uploadError && <div style={{ color: 'red', marginTop: 10 }}>{uploadError}</div>}
      {uploadSuccess && <div style={{ color: 'green', marginTop: 10 }}>{uploadSuccess}</div>}
    </div>
  );
}

export default CollectionDetail;
