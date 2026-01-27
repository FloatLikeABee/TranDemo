import axios from 'axios';

const API_BASE_URL = process.env.REACT_APP_API_URL || 'http://localhost:9090';

// Form Templates API
export const getFormTemplates = async (userType = '') => {
  const url = userType 
    ? `${API_BASE_URL}/api/forms/templates?user_type=${userType}`
    : `${API_BASE_URL}/api/forms/templates`;
  const response = await axios.get(url);
  return response.data;
};

export const getFormTemplate = async (id) => {
  const response = await axios.get(`${API_BASE_URL}/api/forms/templates/${id}`);
  return response.data;
};

export const createFormTemplate = async (template) => {
  const response = await axios.post(`${API_BASE_URL}/api/forms/templates`, template);
  return response.data;
};

export const updateFormTemplate = async (id, template) => {
  const response = await axios.put(`${API_BASE_URL}/api/forms/templates/${id}`, template);
  return response.data;
};

export const deleteFormTemplate = async (id) => {
  const response = await axios.delete(`${API_BASE_URL}/api/forms/templates/${id}`);
  return response.data;
};

// Form Answers API
export const getFormAnswers = async (formId = '', userId = '') => {
  let url = `${API_BASE_URL}/api/forms/answers`;
  const params = [];
  if (formId) params.push(`form_id=${formId}`);
  if (userId) params.push(`user_id=${userId}`);
  if (params.length > 0) url += '?' + params.join('&');
  
  const response = await axios.get(url);
  return response.data;
};

export const getFormAnswer = async (id) => {
  const response = await axios.get(`${API_BASE_URL}/api/forms/answers/${id}`);
  return response.data;
};

export const createFormAnswer = async (answer) => {
  const response = await axios.post(`${API_BASE_URL}/api/forms/answers`, answer);
  return response.data;
};

export const updateFormAnswer = async (id, answer) => {
  const response = await axios.put(`${API_BASE_URL}/api/forms/answers/${id}`, answer);
  return response.data;
};

export const deleteFormAnswer = async (id) => {
  const response = await axios.delete(`${API_BASE_URL}/api/forms/answers/${id}`);
  return response.data;
};
