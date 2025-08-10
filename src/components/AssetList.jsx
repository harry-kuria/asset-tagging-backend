import React, { useState, useEffect } from 'react';
import axios from 'axios';
import { useNavigate } from 'react-router-dom';
import { Table, Button } from 'react-bootstrap';

const AssetList = () => {
  const [assets, setAssets] = useState([]);
  const navigate = useNavigate();

  useEffect(() => {
    const fetchAssets = async () => {
      try {
        const response = await axios.get('http://localhost:5000/api/assets');
        setAssets(response.data);
      } catch (error) {
        console.error('Error fetching assets:', error);
      }
    };

    fetchAssets();
  }, []);

  const handleEdit = (assetId) => {
    // Navigate to the edit asset page with the asset ID
    navigate(`/edit-asset/${assetId}`);
  };

  const handleDelete = async (assetId) => {
    try {
      await axios.delete(`http://localhost:5000/api/assets/${assetId}`);
      // Refresh the asset list after deletion
      const response = await axios.get('http://localhost:5000/api/assets');
      setAssets(response.data);
    } catch (error) {
      console.error('Error deleting asset:', error);
      // Show error message
    }
  };

  return (
    <div>
      <h2>Asset List</h2>
      <Table striped bordered hover>
        <thead>
          <tr>
            <th>Asset Name</th>
            <th>Asset Type</th>
            <th>Location</th>
            <th>Actions</th>
          </tr>
        </thead>
        <tbody>
          {assets.map((asset) => (
            <tr key={asset.id}>
              <td>{asset.assetName}</td>
              <td>{asset.assetType}</td>
              <td>{asset.location}</td>
              <td>
                <Button variant="primary" onClick={() => handleEdit(asset.id)}>
                  Edit
                </Button>{' '}
                <Button variant="danger" onClick={() => handleDelete(asset.id)}>
                  Delete
                </Button>
              </td>
            </tr>
          ))}
        </tbody>
      </Table>
    </div>
  );
};

export default AssetList;
