import React, { useState } from "react";
import { BrowserRouter as Router, Route, Routes } from 'react-router-dom';
import DashBoard from './components/MainPanel';
import Home from './components/Home'
import AddAsset from './components/AddAsset';
import EncodeBarcode from './components/EncodeBarcode';
import ViewReport from './components/ViewReport';
import AddUserForm from './components/AddUserForm';
import Users from './components/UserList';
import EditUser from './components/EditUser';
import AssetList from './components/AssetList';
import EditAsset from './components/EditAsset';
import MultipleEncode from './components/MultipleEncode';


const AppRouter = () => {
  const [DarkTheme, setDarkTheme] = useState(true);
  return (
   
      <Router>
        
        <Routes>
          <Route path='/' element={<Home />} />
          <Route path='/dashboard' element={<DashBoard />} />
          <Route path='/add_asset' element={<AddAsset />} />
          <Route path='/encode' element={<EncodeBarcode />} />
          <Route path='/reports' element={<ViewReport />} />
          <Route path='/adduser' element={<AddUserForm />} />
          <Route path='/users' element={<Users />} />
          <Route path='/assets' element={<AssetList />} />
          <Route path='/encode_multiple' element={<MultipleEncode />} />
          <Route path='/edit-user/:id' element={<EditUser />} />
          <Route path='/edit-asset/:id' element={<EditAsset />} />
        </Routes>
      </Router>
   
  );
}

export default AppRouter;
