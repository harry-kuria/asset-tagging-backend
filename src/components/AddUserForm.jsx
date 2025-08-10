// AddUserForm.jsx

import React, { useState } from 'react';
import {
  TextField,
  Button,
  Container,
  Typography,
  Checkbox,
  FormControlLabel,
  FormGroup,
  FormControl,
} from '@mui/material';
import axios from 'axios';
import { toast, ToastContainer } from 'react-toastify';
import 'react-toastify/dist/ReactToastify.css';

const AddUserForm = () => {
  const [user, setUser] = useState({
    username: '',
    password: '',
    roles: {
      userManagement: false,
      assetManagement: false,
      encodeAssets: false,
      addMultipleAssets: false,
      viewReports: false,
      printReports: false,
      // Add more roles as needed
    },
  });


  const handleInputChange = (e) => {
    if (e.target.type === 'checkbox') {
      setUser({
        ...user,
        roles: {
          ...user.roles,
          [e.target.name]: e.target.checked,
        },
      });
    } else {
      setUser({ ...user, [e.target.name]: e.target.value });
    }
  };

  const handleAddUser = async () => {
    try {
      const response = await axios.post('http://localhost:5000/api/addUser', user);
      console.log(response.data); // Handle the response as needed
      toast.success('User added successfully', {
        position: 'top-right',
        autoClose: 3000, // Close the toast after 3000 milliseconds (3 seconds)
        hideProgressBar: false,
        closeOnClick: true,
        pauseOnHover: true,
        draggable: true,
        progress: undefined,
      });
    } catch (error) {
      console.error('Error adding user:', error);
    }
  };

  return (
    <Container>
       <Typography variant="h4" gutterBottom>
        Add User
      </Typography>
      <form>
        <TextField
          label="Username"
          name="username"
          value={user.username}
          onChange={handleInputChange}
          fullWidth
          margin="normal"
        />
        <TextField
          label="Password"
          name="password"
          type="password"
          value={user.password}
          onChange={handleInputChange}
          fullWidth
          margin="normal"
        />
        <FormControl component="fieldset">
          <Typography variant="h6">Select User Roles:</Typography>
          <FormGroup>
            <FormControlLabel
              control={
                <Checkbox
                  checked={user.roles.userManagement}
                  onChange={handleInputChange}
                  name="userManagement"
                />
              }
              label="Manage Users"
            />
            <FormControlLabel
              control={
                <Checkbox
                  checked={user.roles.assetManagement}
                  onChange={handleInputChange}
                  name="assetManagement"
                />
              }
              label="Manage Assets"
            />
            <FormControlLabel
              control={
                <Checkbox
                  checked={user.roles.encodeAssets}
                  onChange={handleInputChange}
                  name="encodeAssets"
                />
              }
              label="Asset Encoding"
            />
            <FormControlLabel
              control={
                <Checkbox
                  checked={user.roles.addMultipleAssets}
                  onChange={handleInputChange}
                  name="addMultipleAssets"
                />
              }
              label="Add Multiple Assets"
            />
            <FormControlLabel
              control={
                <Checkbox
                  checked={user.roles.viewReports}
                  onChange={handleInputChange}
                  name="viewReports"
                />
              }
              label="View Reports"
            />
            <FormControlLabel
              control={
                <Checkbox
                  checked={user.roles.printReports}
                  onChange={handleInputChange}
                  name="printReports"
                />
              }
              label="Print Reports"
            />
          </FormGroup>
          <Button variant="contained" color="primary" onClick={handleAddUser}>
          Add User
        </Button>
        </FormControl>
       
        <ToastContainer position="top-right" autoClose={3000} />
      </form>
    </Container>
  );
};

export default AddUserForm;
