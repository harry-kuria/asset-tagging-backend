import React, { useState, useEffect, useRef } from 'react';
import { Form, Button, ListGroup, Container, Modal, Dropdown, Card } from 'react-bootstrap';
import AssetDetails from './AssetDetails';
import axios from 'axios';
import html2canvas from 'html2canvas';
import Barcode from 'react-barcode';
import ReactToPdf from 'react-to-pdf';
import jsPDF from 'jspdf'; // Import jsPDF library
import 'jspdf-autotable';

const MultipleEncode = () => {
  const [searchTerm, setSearchTerm] = useState('');
  const [searchResults, setSearchResults] = useState([]);
  const [assetDetails, setAssetDetails] = useState(null);
  const [barcodeType, setBarcodeType] = useState(null);
  const [barcodeData, setBarcodeData] = useState(null);
  const [showAssetModal, setShowAssetModal] = useState(false);
  const [departments, setDepartments] = useState([]);
  const [functionalAreas, setFunctionalAreas] = useState([]);

  const barcodesComponentRef = useRef();
  const handleDownloadPdf = () => {
    try {
      const pdf = new jsPDF();
  
      pdf.text('Generated Barcodes', 20, 20);
  
      // Define the table headers
      const headers = [['Asset Name', 'Code', 'Location']];
  
      // Extract table rows from generatedBarcodes
      const data = generatedBarcodes.map((barcode) => [
        barcode.assetDetails.assetName,
        barcode.formattedString,
        barcode.assetDetails.location,
      ]);
  
      // Add the table with autoTable
      pdf.autoTable({
        startY: 30, // Adjust the starting position as needed
        head: headers,
        body: data,
        theme: 'grid', // You can choose a different theme if needed
        headStyles: {
          fillColor: [12, 83, 245]
          }
      });
  
      pdf.save('generated_barcodes.pdf');
    } catch (error) {
      console.error('Error capturing barcodes to PDF:', error);
    }
  };

  const [filters, setFilters] = useState({
    institution: '',
    department: '',
    functionalArea: '',
  });
  const [institutionList, setInstitutionList] = useState([]);
  const [generatedBarcodes, setGeneratedBarcodes] = useState([]);
  const handleDownloadBarcode = async (barcode) => {
    try {
      const canvas = await html2canvas(/* Ref of the barcode element */);
      const dataUrl = canvas.toDataURL('image/png');
  
      const a = document.createElement('a');
      a.href = dataUrl;
      a.download = `barcode_${barcode}.png`;
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
    } catch (error) {
      console.error('Error capturing barcode:', error);
    }
  };
  const handleGenerateBarcodes = async () => {
    // Fetch assets based on filters to generate barcodes
    try {
      const response = await axios.post('http://localhost:5000/api/generateBarcodesByInstitutionAndDepartment', {
        institution: filters.institution || null,
        department: filters.department || null,
        functionalArea: filters.functionalArea || null,
      });
  
      const { barcodeTags, assetDetails } = response.data;
  
      // Now you can use barcodeTags and assetDetails as needed
      console.log('Barcode Tags:', barcodeTags);
      console.log('Asset Details:', assetDetails);
  
      setGeneratedBarcodes(barcodeTags);
      setAssetDetails(assetDetails); // Assuming you have a state variable to store asset details
    } catch (error) {
      console.error('Error generating barcodes:', error);
    }
  };

  useEffect(() => {
    const fetchInstitutions = async () => {
      try {
        const response = await axios.get('http://localhost:5000/api/institutions');
        setInstitutionList(response.data);
      } catch (error) {
        console.error('Error fetching institutions:', error);
      }
    };
    const fetchDepartments = async () => {
        try {
          const response = await axios.get('http://localhost:5000/api/departments');
          setDepartments(response.data);
        } catch (error) {
          console.error('Error fetching departments:', error);
        }
      };
  
      const fetchFunctionalAreas = async () => {
        try {
          const response = await axios.get('http://localhost:5000/api/functionalAreas');
          setFunctionalAreas(response.data);
        } catch (error) {
          console.error('Error fetching functional areas:', error);
        }
      };
  
      fetchInstitutions();
      fetchDepartments();
      fetchFunctionalAreas();

    fetchInstitutions();
  }, []);

  useEffect(() => {
    const fetchSearchResults = async () => {
      try {
        const response = await axios.get(`http://localhost:5000/api/searchAssets`, {
          params: {
            term: searchTerm,
            ...filters,
          },
        });
        setSearchResults(response.data);
      } catch (error) {
        console.error('Error fetching search results:', error);
      }
    };

    if (searchTerm.trim() !== '') {
      fetchSearchResults();
    } else {
      setSearchResults([]);
    }
  }, [searchTerm, filters]);

  const handleEncode = async (selectedBarcodeType) => {
    setBarcodeType(selectedBarcodeType);

    if (assetDetails) {
      try {
        const response = await axios.get(`http://localhost:5000/api/getAssetDetails?id=${assetDetails.id}`);
        const detailedInfo = response.data;
        const institutionName = detailedInfo.institutionName.toUpperCase();
        const institutionShort = detailedInfo.institutionName.substring(0, 2).toUpperCase();
        const departmentShort = detailedInfo.department.substring(0, 2).toUpperCase();
        const functionalAreaShort = detailedInfo.functionalArea.substring(0, 2).toUpperCase();
        const assetNameShort = detailedInfo.assetName.replace(/\s/g, '').toUpperCase();
        const id = detailedInfo.id;

        const dataForBarcode = `${institutionShort}/${departmentShort}/${functionalAreaShort}/${assetNameShort}/${id}`;
        setBarcodeData(dataForBarcode);
      } catch (error) {
        console.error('Error fetching asset details:', error);
      }
    }
  };

  const barcodeRef = useRef(null);

  const handleDownload = async () => {
    try {
      const canvas = await html2canvas(barcodeRef.current);
      const dataUrl = canvas.toDataURL('image/png');

      const a = document.createElement('a');
      a.href = dataUrl;
      a.download = 'barcode.png';
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
    } catch (error) {
      console.error('Error capturing barcode:', error);
    }
  };

  useEffect(() => {
    // Implement barcode generation logic here when barcodeType changes
  }, [barcodeType]);

  const handleSelectAsset = (selectedAsset) => {
    setAssetDetails(selectedAsset);
    setSearchTerm('');
    setSearchResults([]);
    setShowAssetModal(true);
  };

  const handleCloseAssetModal = () => {
    setShowAssetModal(false);
  };

  return (
    <Container className="mt-5">
      <h2 className="mb-4">View Assets</h2>

      <Form>
      <Form.Group controlId="institution">
        <Form.Label>Institution</Form.Label>
        <Dropdown>
          <Dropdown.Toggle variant="success" id="dropdown-basic">
            {filters.institution ? filters.institution : 'All'}
          </Dropdown.Toggle>

          <Dropdown.Menu>
            <Dropdown.Item onClick={() => setFilters({ ...filters, institution: null })}>
              All
            </Dropdown.Item>
            {institutionList.map((institution) => (
              <Dropdown.Item
                key={institution}
                onClick={() => setFilters({ ...filters, institution })}
              >
                {institution}
              </Dropdown.Item>
            ))}
          </Dropdown.Menu>
        </Dropdown>
      </Form.Group>

      <Form.Group controlId="department">
        <Form.Label>Department</Form.Label>
        <Dropdown>
          <Dropdown.Toggle variant="success" id="dropdown-basic">
            {filters.department ? filters.department : 'All'}
          </Dropdown.Toggle>

          <Dropdown.Menu>
            <Dropdown.Item onClick={() => setFilters({ ...filters, department: null })}>
              All
            </Dropdown.Item>
            {departments.map((department) => (
              <Dropdown.Item
                key={department}
                onClick={() => setFilters({ ...filters, department })}
              >
                {department}
              </Dropdown.Item>
            ))}
          </Dropdown.Menu>
        </Dropdown>
      </Form.Group>
     
        <Form.Group controlId="searchTerm">
          <Form.Label>Search Asset</Form.Label>
          <Form.Control
            type="text"
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            placeholder="Enter asset name or ID"
          />
        </Form.Group>
      </Form>

      {searchResults.length > 0 && (
        <ListGroup>
          {searchResults.map((result) => (
            <ListGroup.Item
              key={result.id}
              action
              onClick={() => handleSelectAsset(result)}
              className="cursor-pointer"
            >
              {result.assetName}
            </ListGroup.Item>
          ))}
        </ListGroup>
      )}

<Button variant="primary" onClick={handleGenerateBarcodes}>
        Generate Barcodes
      </Button>

      {generatedBarcodes.length > 0 && (
         <div className="mt-4" ref={barcodesComponentRef}>
          <h3>Generated Barcodes</h3>
          <Card>
            <Card.Body>
            <table className="table">
          <thead>
            <tr>
              <th>Name</th>
              <th>Code</th>
              <th>Location</th>
              
            </tr>
          </thead>
          <tbody>
            {generatedBarcodes.map((barcode, index) => (
              <tr key={index}>
                <td>{barcode.assetDetails.assetName}</td>
                <td>{barcode.formattedString}</td>
                <td>{barcode.assetDetails.location}</td>
                
              </tr>
            ))}
          </tbody>
        </table>
              {/* {generatedBarcodes.map((barcode, index) => (
                <div key={index} className="mb-3">
                  <div style={{ display: 'flex', alignItems: 'center', gap: '10px' }}>
                    <div style={{ width: '50px' }}>
                      
                      <Barcode value={barcode.barcodeTag} /> 
                    </div>
                    <div>
                      <p>{barcode.formattedString}</p>
                      
                      <table>
                        <tbody>
                          <tr>
                            <td>Asset Name:</td>
                            <td>{barcode.assetDetails.assetName}</td>
                          </tr>
                          <tr>
                            <td>Department:</td>
                            <td>{barcode.assetDetails.department}</td>
                          </tr>
                          <tr>
                            <td>Location:</td>
                            <td>{barcode.assetDetails.location}</td>
                          </tr>
                         
                        </tbody>
                      </table>
                      <Button variant="secondary" onClick={() => handleDownloadBarcode(barcode.barcodeTag)}>
                        Download Barcode
                      </Button>
                    </div>
                  </div>
                </div>
              ))} */}
              
            </Card.Body>
          </Card>
        </div>
      )}
      <Button variant="primary" onClick={() => handleDownloadPdf()}>
          Export Barcodes to PDF
        </Button>
           
      {/* Add the button for exporting to PDF */}
      {/* <Button variant="primary" onClick={toPdf}>
        Export Barcodes to PDF
      </Button> */}


      <Modal show={showAssetModal} onHide={handleCloseAssetModal} size="lg">
        <Modal.Header closeButton>
          <Modal.Title>Asset Details</Modal.Title>
        </Modal.Header>
        <Modal.Body>
          {assetDetails && <AssetDetails details={assetDetails} />}
        </Modal.Body>
      </Modal>
    </Container>
  );
};

export default MultipleEncode;
