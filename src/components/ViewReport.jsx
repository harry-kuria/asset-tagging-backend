// ViewReport.jsx
import React, { useState, useEffect } from 'react';
import { Button, TextField, Grid, Typography, Container, FormControl, InputLabel, Select, MenuItem } from '@mui/material';
import { DatePicker } from '@mui/lab';
import { Box } from '@mui/system';
import axios from 'axios';
import * as XLSX from 'xlsx';
import jsPDF from 'jspdf';
import { degrees, PDFDocument, rgb, StandardFonts } from 'pdf-lib';


const ViewReport = () => {
  const [filters, setFilters] = useState({
    assetType: '',
    location: '',
    status: '',
    startDate: null,
    endDate: null,
    manufacturer: '',
    modelNumber: '',
    institutionName: '',
    department: '',
    functionalArea: '',
  });

  const [assetCategories, setAssetCategories] = useState([]);
  const [manufacturers, setManufacturers] = useState([]);
  const [selectedAssetType, setSelectedAssetType] = useState('');
  const [institutionList, setInstitutionList] = useState([]);
  const [departments, setDepartments] = useState([]);
  const [functionalAreas, setFunctionalAreas] = useState([]);

  const [reportFormat, setReportFormat] = useState('pdf');

  const groupAssetsByDepartment = (assets) => {
    const groupedAssets = {};
  
    for (const asset of assets) {
      const department = asset.department || 'Unknown';
  
      if (!groupedAssets[department]) {
        groupedAssets[department] = [];
      }
  
      groupedAssets[department].push(asset);
    }
  
    return groupedAssets;
  };
  const handleGenerateReport = async () => {
    try {
      const response = await axios.get('http://localhost:5000/api/generateReport', {
        params: {
          assetType: selectedAssetType,
          location: filters.location,
          status: filters.status,
          startDate: filters.startDate?.toISOString(),
          endDate: filters.endDate?.toISOString(),
          manufacturer: filters.manufacturer,
          modelNumber: filters.modelNumber,
          institutionName: filters.institutionName,
          department: filters.department,
          functionalArea: filters.functionalArea,
        },
      });
  
      if (Array.isArray(response.data)) {
        // Organize assets by asset type
        const assetsByType = groupAssetsByType(response.data);
  
        console.log('Report data:', assetsByType);
  
        if (reportFormat === 'pdf') {
          generatePDFReport(assetsByType);
        } else if (reportFormat === 'excel') {
          generateExcelFile(assetsByType);
        }
      } else {
        console.error('Unexpected response format:', response.data);
      }
    } catch (error) {
      console.error('Error generating report:', error);
    }
  };
  
  // Helper function to group assets by assetType
  const groupAssetsByType = (assets) => {
    const groupedAssets = {};
  
    for (const asset of assets) {
      const assetType = (asset.assetType || 'Unknown').toLowerCase();
  
      if (!groupedAssets[assetType]) {
        groupedAssets[assetType] = [];
      }
  
      groupedAssets[assetType].push(asset);
    }
  
    return groupedAssets;
  };


  const generateExcelFile = async (reportData) => {
    try {
      // Fetch assets associated with the selected institution
      const assetResponse = await axios.post('http://localhost:5000/api/fetchAssetsByInstitution', {
        institutionName: filters.institutionName,
      });
      const { success, assets } = assetResponse.data;
  
      if (!success) {
        console.error('Error fetching assets:', assets);
        return;
      }
  
      const wb = XLSX.utils.book_new();
  
      // Handle different data formats
      if (!reportData || typeof reportData !== 'object') {
        console.error('Unexpected response format:', reportData);
        return;
      }
  
      // Filter reportData based on the selected institution
      const filteredReportData = {};
      for (const assetType in reportData) {
        if (reportData.hasOwnProperty(assetType)) {
          const assetsOfType = reportData[assetType].filter(asset => asset.institutionName === filters.institutionName);
          if (assetsOfType.length > 0) {
            filteredReportData[assetType] = assetsOfType;
          }
        }
      }
  
      // Generate summary sheet
      const summaryHeaders = ['Asset Type', 'Total Market Value (Kshs.)'];
      const summaryData = [];
  
      for (const assetType in filteredReportData) {
        if (filteredReportData.hasOwnProperty(assetType)) {
          const assets = filteredReportData[assetType];
          const totalMarketValue = assets.reduce((sum, asset) => sum + (asset.marketValue || 0), 0);
          summaryData.push([assetType, totalMarketValue.toFixed(2)]);
        }
      }
  
      const summaryWs = XLSX.utils.json_to_sheet([summaryHeaders, ...summaryData]);
      XLSX.utils.book_append_sheet(wb, summaryWs, 'Summary');
  
      // Generate sheets for each asset type
      for (const assetType in filteredReportData) {
        if (filteredReportData.hasOwnProperty(assetType)) {
          const assets = filteredReportData[assetType];
  
          console.log(`Asset Type: ${assetType}, Number of Assets: ${assets.length}`);
  
          // Dynamic headers based on available properties
          const headers = Object.keys(assets[0]);
  
          // Filter out columns with empty headers
          const nonEmptyHeaders = headers.filter(header => header.trim() !== '');
  
          // Filter out columns with no data
          const nonEmptyData = assets.map((asset) => {
            return nonEmptyHeaders.map((header) => asset[header]);
          });
  
          // Filter out columns that are entirely empty
          const nonEmptyColumns = nonEmptyHeaders.filter((_, index) => nonEmptyData.some(row => row[index] !== undefined && row[index] !== null && row[index] !== ''));
  
          // Ensure the header row is added separately
          const assetClassWsData = [nonEmptyColumns, ...nonEmptyData.map(row => nonEmptyColumns.map(header => row[nonEmptyHeaders.indexOf(header)]))];
          const assetClassWs = XLSX.utils.aoa_to_sheet(assetClassWsData);
          XLSX.utils.book_append_sheet(wb, assetClassWs, assetType);
        }
      }
  
      if (wb.SheetNames.length === 0) {
        console.error('Workbook is empty. Check the structure of reportData.');
        return;
      }
  
      XLSX.writeFile(wb, 'asset_report.xlsx');
    } catch (error) {
      console.error('Error generating Excel file:', error);
    }
  };
  
  
  
  
  
  
  
  
  
  


  useEffect(() => {
    const fetchAssetCategories = async () => {
      try {
        const response = await axios.get('http://localhost:5000/api/categories');
        setAssetCategories(response.data);
      } catch (error) {
        console.error('Error fetching asset categories:', error);
      }
    };

    fetchAssetCategories();
  }, []);

  useEffect(() => {
    const fetchManufacturerCategories = async () => {
      try {
        const manufacturerResponse = await axios.get('http://localhost:5000/api/manufacturers');
        setManufacturers(manufacturerResponse.data);
      } catch (error) {
        console.error('Error fetching manufacturer categories:', error);
      }
    };

    fetchManufacturerCategories();
  }, []);
  

  const handleFilterChange = (field, value) => {
    setFilters((prevFilters) => ({ ...prevFilters, [field]: value }));
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


  const generatePDFReport = async () => {
    try {
      // Fetch assets associated with the selected institution
      const assetResponse = await axios.post('http://localhost:5000/api/fetchAssetsByInstitution', {
        institutionName: filters.institutionName,
      });
  
      const { success, assets } = assetResponse.data;
  
      if (!success) {
        console.error('Error fetching assets:', assets);
        return;
      }
  
      // Group assets by assetType
      const assetsByType = groupAssetsByType(assets);
  
      // Create a new jsPDF instance
      const pdf = new jsPDF({
        orientation: 'landscape',
      });
  
      // Add title for the summary table
      pdf.setFontSize(18);
      pdf.text(`Asset Register of ${filters.institutionName}`, 10, 10);
  
      // Create a summary table with columns for each asset type
      
      const customOrder = ['land','building', 'computer','furniture','electronics', 'motor vehicle', 'motor cycle'];
      const summaryHeaders = ['Asset Type', 'Total Market Value (Kshs.)'];
      const summaryData = [];
      let grandTotalMarketValue = 0; // Initialize grand total
      // Iterate through the custom order
      for (const assetType of customOrder) {
        if (assetsByType.hasOwnProperty(assetType)) {
          // Calculate total market value for the asset type
          let totalMarketValue = 0;
          
      
          // Sum up market value for non-land assets
          if (assetType !== 'land') {
            totalMarketValue = assetsByType[assetType].reduce((sum, asset) => sum + (parseFloat(asset.marketValue) || 0), 0);
          }
      
          // Add disposal value for land assets
          if (assetType === 'land') {
            totalMarketValue = assetsByType[assetType].reduce((sum, asset) => sum + (parseFloat(asset.disposal_value) || 0), 0);
          }
          
          if (assetType === 'building') {
            totalMarketValue = assetsByType[assetType].reduce((sum, asset) => sum + (parseFloat(asset.valuation) || 0), 0);
          }
          if (assetType === 'motorvehicle') {
            totalMarketValue = assetsByType[assetType].reduce((sum, asset) => sum + (parseFloat(asset.amount) || 0), 0);
          }
          if (assetType === 'motor cycle') {
            totalMarketValue = assetsByType[assetType].reduce((sum, asset) => sum + (parseFloat(asset.amount) || 0), 0);
          }
      
          // Add data to the summary table
          summaryData.push([assetType.toUpperCase(), totalMarketValue.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })]);
         // Accumulate grand total
    grandTotalMarketValue += totalMarketValue;
  }
}

// Add a row for the grand total
summaryData.push([
  { content: 'Grand Total', styles: { fontStyle: 'bold', textColor: [0, 0, 0] } },  // Bold and black text
  { content: grandTotalMarketValue.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 }), styles: { fontStyle: 'bold', textColor: [0, 0, 0] } },  // Bold and black text
]);     
  
      // AutoTable to generate the summary table on the first page
      pdf.autoTable({
        head: [[{ content: 'Asset Type', styles: { fontStyle: 'bold',  fontSize: 10, align: 'left' } }, { content: 'Total Market Value (Kshs.)', styles: { fontStyle: 'bold', cellPadding: 1, lineHeight: 10, halign: 'right', fontSize: 10 } }]],

        body: summaryData,
        startY: 30, // Adjust startY as needed
        margin: { top: 30 },
        bodyStyles: { fontStyle: 'normal', textColor: [0, 0, 0], fontSize: 10 },  // Default style for the table body
        columnStyles: {
          0: { fontStyle: 'normal', textColor: [0, 0, 0], cellWidth: 'auto', align: 'left' },  // Style for the first column
          1: { fontStyle: 'normal', textColor: [0, 0, 0], cellPadding: 1, lineHeight: 10, halign: 'right' },  // Style for the second column (total market value)
        },
        //styles: { cellPadding: 1, lineHeight: 10, halign: 'right' },
      });
      const currentDate = new Date();

      // Format the date as needed
      const formattedDate = `${currentDate.getDate()}/${currentDate.getMonth() + 1}/${currentDate.getFullYear()}`;
      pdf.setFontSize(15);
      const preparedByText = 'Prepared by\nMoowi Company Group Limited\nP.o Box 13977-00400\nNairobi\n';
      const textWidth = pdf.getStringUnitWidth(preparedByText) * pdf.internal.getFontSize() / pdf.internal.scaleFactor;
      const textX = (pdf.internal.pageSize.width ) / 2;
      pdf.text(preparedByText + formattedDate, textX, pdf.lastAutoTable.finalY + 20, { align: 'center' });  // Center-justify the text
      
  
      // Add a new page for each asset type
      for (const assetType in assetsByType) {
        if (assetsByType.hasOwnProperty(assetType)) {
          // Add a new page for each assetType
          pdf.addPage();
  
          // Add letterhead or any content for each page if needed
          pdf.setFontSize(10);
          pdf.text(`${assetType}`, 10, 10);
  
          // Table headers
          const headers = ['Asset Name', 'Tag No.', 'Class', 'Location', 'Purchase Price (Kshs.)', 'Market Value/Revalued Amount', 'Model Type', 'Department', 'Functional Area', 'Serial Number','Vehicle Reg No',	'Source of funds',	'Engine Number',	'Chassis Number',	'Make',	'Purchase year',	'PVNumber',	'Original Location',	'Current Location',	'Replacement Date',	'Amount',	'Depreciation Rate',	'Annual Depreciation',	'Accumulated Depreciation',	'Netbook Value',	'Disposal Date',	'Responsible Officer',	'Asset Condition', 'Description','Ownership','Institution Number','Nearest Town','Street','County','LR Certificate Number','Size','Ownership Status','Mode of Acquisition','Building Type','Designated Use','Number of floors','Area','Valuation','Annual Depreceation','Estimated Useful Life','Accumulated Depreceation Rate','Netbook Value','Annual Rental Income','Land Description','Mode of Acquisition','County','Nearest Town','GPS','Polygon','LR Certificate No.','Ownership Document','Ownership Details','Size','Ownership Status','Acquisition Date','Registration Date','Disputed/Undisputed','Encumberances','Planned/Unplanned','Purpose','Surveyed/Not Surveyed','Acquisition Amount','Fair Value','Disposal Date','Disposal Value','Annual Rental Income of Land'
          ];
  
          // Table data
          const data = assetsByType[assetType].map((asset) => [
            asset.assetName || '',
            asset.formattedString || '',
            asset.assetType || '',
            asset.location || '',
            asset.purchasePrice || '',
            asset.marketValue || '',
            asset.manufacturer || '',
            asset.department || '',
            asset.functionalArea || '',
            asset.serialNumber || '',
            asset.vehicleregno || '',
            asset.sourceoffunds || '',
            asset.enginenumber || '',	
            asset.chassisnumber || '',
            asset.make || '',	
            asset.purchaseyear || '',
            asset.pvnumber || '',	
            asset.originallocation || '',
            asset.currentlocation || '',
            asset.replacementdate || '',
            asset.amount || '',
            asset.depreciationrate || '',
            asset.annualdepreciation || '',
            asset.accumulateddepreciation || '',
            asset.netbookvalue || '',	
            asset.disposaldate || '',	
            asset.responsibleofficer || '',
            asset.assetcondition || '',
            asset.description || '',
            asset.ownership || '',
            asset.institutionno || '',
            asset.nearesttown || '',
            asset.street || '',
            asset.county || '',
            asset.lrno || '',
            asset.sizeofland || '',
            asset.ownershipstatus || '',
            asset.sourceoffunds || '',
            asset.modeofacquisition || '',
            asset.buildingtype || '',
            asset.designateduse || '',
            asset.nooffloors || '',
            asset.area || '',
            asset.valuation || '',
            asset.annualdepreciation || '',
            asset.estimatedusefullife || '',
            asset.accumulateddepreciationrate || '',
            asset.netbookvalue || '',
            asset.annualrentalincome || '',
            asset.land_description || '',
            asset.mode_of_acquisition || '',
            asset.nearest_town_location || '',
            asset.gps || '',
            asset.polygon || '',
            asset.lr_certificate_no || '',
            asset.document_of_ownership || '',
            asset.ownership_details || '',
            asset.size_ha || '',
            asset.ownership_status || '',
            asset.acquisition_date || '',
            asset.registration_date || '',
            asset.disputed_undisputed || '',
            asset.encumbrances || '',
            asset.planned_unplanned || '',
            asset.purpose_use_of_land || '',
            asset.surveyed_not_surveyed || '',
            asset.acquisition_amount || '',
            asset.fair_value_ministry_of_lands || '',
            asset.disposal_date_change_of_use_date || '',
            asset.disposal_value || '',
            asset.annual_rental_income || ''
          ]);
  
          // Filter out columns with all empty values
          const filteredHeaders = headers.filter((_, index) => data.some((row) => row[index] !== ''));
          const filteredData = data.map((row) => row.filter((value) => value !== ''));
          const columnWidths = {
            0: 16,  // Asset Name
            1: 20,  // Tag No.
            2: 10,  // Class
            3: 10,  // Location
            4: 10,  // Purchase Price (Kshs.)
            5: 10,  // Market Value/Revalued Amount
            // ... set widths for other columns
          };
      
          // Define styles for each column
          const columnStyles = {};
          filteredHeaders.forEach((header, index) => {
            columnStyles[index] = { cellWidth: columnWidths[index] || 'auto' };
          });
      
          // AutoTable to generate the table
          pdf.autoTable({
            head: [filteredHeaders],
            body: filteredData,
            startY: 45, // Adjust as needed
            margin: { top: 45 }, // Adjust top margin for the table
            styles: { fontSize: 8 }, // Set the font size for the table content
            columnStyles: columnStyles, // Set the column widths
          });
        }
      }
  
      // Save the PDF
      pdf.save('asset_register.pdf');
    } catch (error) {
      console.error('Error generating asset register report:', error);
    }
  };
  
  const handleGenerateCertificate = async () => {
    try {
      // Fetch assets associated with the selected institution
      const assetResponse = await axios.post('http://localhost:5000/api/fetchAssetsByInstitution', {
        institutionName: filters.institutionName,
      });
  
      const { success, assets } = assetResponse.data;
  
      if (!success) {
        console.error('Error fetching assets:', assets);
        return;
      }
  
      // Group assets by assetType
      const assetsByType = groupAssetsByType(assets);
  
      // Calculate totals for both "Land" and "Buildings" asset types
      const landTotal = calculateAssetTypeTotal(assetsByType, 'land', 'disposal_value');
      const buildingTotal = calculateAssetTypeTotal(assetsByType, 'building', 'valuation');
      const furnitureTotal = calculateAssetTypeTotal(assetsByType, 'furniture', 'marketValue');
      const computerTotal = calculateAssetTypeTotal(assetsByType, 'computer', 'marketValue');
      const electronicTotal = calculateAssetTypeTotal(assetsByType, 'electronics', 'marketValue');
      const mvTotal = calculateAssetTypeTotal(assetsByType, 'motorvehicle', 'amount');
      const grandTotal = landTotal + buildingTotal + furnitureTotal + computerTotal + mvTotal;
      console.log('Land Total:', landTotal.toLocaleString());
      console.log('Building Total:', buildingTotal.toLocaleString());
      const totalLandAndBuildings = landTotal + buildingTotal;
      const fCEBIM = furnitureTotal + computerTotal + electronicTotal + 0 + 0 + mvTotal;
      const institution = filters.institutionName;
      // Send a POST request to the generateInvoice endpoint with both totals
      const generateInvoiceResponse = await axios.post('http://localhost:5000/api/generate_invoice', {
        landTotal, buildingTotal,furnitureTotal,computerTotal,mvTotal,electronicTotal,grandTotal,totalLandAndBuildings,institution,fCEBIM });
        console.log('Before fetch');
        try {
          const downloadResponse = await axios.get('http://localhost:5000/api/download');
          console.log('Download Response:', downloadResponse);
        
          // Assuming the response contains the data you need
          const blob = new Blob([downloadResponse.data], { type: 'application/pdf' });
        
          // Create a link and simulate a click to trigger the download
          const url = window.URL.createObjectURL(blob);
          const a = document.createElement('a');
          a.href = url;
          a.download = 'Moowi_Valuation_Certificate.pdf';
          document.body.appendChild(a);
          a.click();
          document.body.removeChild(a);
        } catch (error) {
          console.error('Error downloading:', error);
        }
    } catch (error) {
      console.error('Error generating certificate:', error);
    }
  };
  
  // Helper function to calculate the total for a specific asset type
  const calculateAssetTypeTotal = (assetsByType, assetType, property) => {
  const assetTypeAssets = assetsByType[assetType] || [];
  const assetTypeTotal = assetTypeAssets.reduce((sum, asset) => {
    const propertyValue = asset[property];
    const trimmedValue = String(propertyValue).trim(); // Convert to string and trim
    const parsedValue = parseFloat(trimmedValue) || 0;
    return sum + parsedValue;
  }, 0);

  return assetTypeTotal;
};
  
  
  
  

  const calculateTotals = (assetsByType) => {
    const totals = {
      grandTotal: 0,
      assetTypes: {},
    };
  
    for (const assetType in assetsByType) {
      if (assetsByType.hasOwnProperty(assetType)) {
        const assets = assetsByType[assetType];
        const totalMarketValue = assets.reduce((sum, asset) => sum + (parseFloat(asset.marketValue) || 0), 0);
  
        totals.assetTypes[assetType] = {
          totalMarketValue: totalMarketValue.toFixed(2),
        };
  
        totals.grandTotal += totalMarketValue;
      }
    }
  
    return totals;
  };
  
  
  
  
  // Helper function to group assets by assetType
  
  
  
  
  // ...

  return (
  <Container>
    <Typography variant="h4" gutterBottom>
      Asset Register Report
    </Typography>
    <form>
    <Grid container spacing={2} justifyContent="center">
          <Grid item xs={12} sm={6}>
            <FormControl fullWidth variant="outlined">
              <InputLabel id="institutionNameLabel">Institution Name</InputLabel>
              <Select
                labelId="institutionNameLabel"
                id="institutionName"
                value={filters.institutionName}
                onChange={(e) => handleFilterChange('institutionName', e.target.value)}
                label="Institution Name"
              >
                <MenuItem value="">All</MenuItem>
                {institutionList.map((institution) => (
                  <MenuItem key={institution} value={institution}>
                    {institution}
                  </MenuItem>
                ))}
              </Select>
            </FormControl>
          </Grid>
        </Grid>
      <Box sx={{ display: 'flex', justifyContent: 'center', marginTop: 2 }}> {/* Center the button */}
      

          

      <Button type="button" variant="contained" color="primary" onClick={handleGenerateReport}>
            Generate Report
          </Button>
          <FormControl fullWidth variant="outlined" sx={{ marginLeft: 2 }}>
            <InputLabel id="reportFormatLabel">Report Format</InputLabel>
            <Select
              labelId="reportFormatLabel"
              id="reportFormat"
              value={reportFormat}
              onChange={(e) => setReportFormat(e.target.value)}
              label="Report Format"
            >
              <MenuItem value="pdf">PDF</MenuItem>
              <MenuItem value="excel">Excel</MenuItem>
            </Select>
          </FormControl>
          <Button
            type="button"
            variant="contained"
            color="secondary"
            onClick={handleGenerateCertificate}
            sx={{ marginLeft: 2 }}
          >
            Generate Certificate
          </Button>
      </Box>
    </form>
  </Container>
);
};

export default ViewReport;
