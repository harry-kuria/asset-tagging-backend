import React from 'react';
import { Container, Row, Col, Nav, Button, Card } from 'react-bootstrap';
import { useNavigate } from 'react-router-dom';


const CustomCard = ({ title, color,text, handleClick }) => (
  
  <Card bg={color} text="white" style={{ width: '18rem', margin: '1rem' }}>
    <Card.Header>{title}</Card.Header>
    <Card.Body>
      <Card.Text>
        {text}
      </Card.Text>
      <Button variant="light" onClick={handleClick}>Go to {title}</Button>
    </Card.Body>
  </Card>
);
const MainPanel = () => {
  const navigate = useNavigate();
    const navLinkStyle = {
        background: '#0b5cd2',
        color: 'white',
        padding: '0.5rem',
        borderRadius: '5px',
        margin: '0.5rem 0',
      };
  return (
    <Container fluid>
      <Row>
        {/* Left Menu */}
        <Col sm={12} md={3} lg={2} className="bg-light">
          <Nav className="flex-column">
            {/* User Section */}
            <Nav.Link href="#user-section" style={navLinkStyle}>User Section</Nav.Link>
            <Nav.Link href="/users">View User</Nav.Link>
            
            <Nav.Link href="/adduser">Create User</Nav.Link>

            {/* Asset Section */}
            <Nav.Link href="#asset-section" style={navLinkStyle}>Asset Section</Nav.Link>
            <Nav.Link href="/add_asset">Create Asset</Nav.Link>
            <Nav.Link href="/assets">View Assets</Nav.Link>
            

            {/* Barcode Section */}
            <Nav.Link href="#barcode-section" style={navLinkStyle}>Barcode Section</Nav.Link>
            <Nav.Link href="/encode">Encode Barcode</Nav.Link>

            {/* Report Section */}
            <Nav.Link href="#report-section" style={navLinkStyle}>Report Section</Nav.Link>
            <Nav.Link href="/reports">View Report</Nav.Link>
          </Nav>
        </Col>

        {/* Content Area */}
        <Col sm={12} md={9} lg={10} className="p-4">
         

          {/* Example Cards */}
          <Row>
            <Col>
            <CustomCard title="Add Multiple Assets" color="primary" text="Upload an Excel document with Multiple assets details as per the System Format" handleClick={() => navigate('/add_asset')} />
            </Col>
            <Col>
            <CustomCard title="Encode Multiple Assets" color="info" text="Upload an Excel document with Multiple assets details to be encoded" handleClick={() => navigate('/encode_multiple')} />
            </Col>
            <Col>
              <CustomCard title="Encode Assets" color="success" text="Generate barcodes for assets either in Excel or in PDF format" handleClick={() => navigate('/encode')} />
            </Col>
            <Col>
              <CustomCard title="User Management" color="primary" text="View, Edit, delete and create users" handleClick={() => navigate('/adduser')} />
            </Col>
            
          </Row>
        </Col>
      </Row>
    </Container>
  );
};

export default MainPanel;
