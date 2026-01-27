import React, { useState, useEffect } from 'react';
import {
  getFormTemplates,
  getFormTemplate,
  createFormTemplate,
  updateFormTemplate,
  deleteFormTemplate
} from '../../services/formsApi';
import './Forms.css';

const Forms = () => {
  const [forms, setForms] = useState([]);
  const [filteredForms, setFilteredForms] = useState([]);
  const [loading, setLoading] = useState(true);
  const [userTypeFilter, setUserTypeFilter] = useState('');
  const [showModal, setShowModal] = useState(false);
  const [editingForm, setEditingForm] = useState(null);
  const [alert, setAlert] = useState(null);
  const [formData, setFormData] = useState({
    name: '',
    description: '',
    user_type: '',
    fields: []
  });

  useEffect(() => {
    loadForms();
  }, []);

  useEffect(() => {
    filterForms();
  }, [forms, userTypeFilter]);

  const loadForms = async () => {
    try {
      setLoading(true);
      const data = await getFormTemplates();
      setForms(Array.isArray(data) ? data : []);
    } catch (error) {
      showAlert('Error loading forms: ' + (error.response?.data?.error || error.message), 'error');
      setForms([]);
    } finally {
      setLoading(false);
    }
  };

  const filterForms = () => {
    const list = Array.isArray(forms) ? forms : [];
    if (!userTypeFilter) {
      setFilteredForms(list);
    } else {
      setFilteredForms(list.filter(f => f.user_type === userTypeFilter));
    }
  };

  const showAlert = (message, type) => {
    setAlert({ message, type });
    setTimeout(() => setAlert(null), 5000);
  };

  const openCreateModal = () => {
    setEditingForm(null);
    setFormData({
      name: '',
      description: '',
      user_type: '',
      fields: []
    });
    setShowModal(true);
  };

  const openEditModal = async (id) => {
    try {
      const form = await getFormTemplate(id);
      setEditingForm(id);
      setFormData({
        name: form.name,
        description: form.description || '',
        user_type: form.user_type,
        fields: form.fields || []
      });
      setShowModal(true);
    } catch (error) {
      showAlert('Error loading form: ' + (error.response?.data?.error || error.message), 'error');
    }
  };

  const closeModal = () => {
    setShowModal(false);
    setEditingForm(null);
    setFormData({
      name: '',
      description: '',
      user_type: '',
      fields: []
    });
  };

  const addField = () => {
    setFormData({
      ...formData,
      fields: [...formData.fields, {
        name: '',
        label: '',
        type: 'text',
        required: false,
        placeholder: '',
        options: []
      }]
    });
  };

  const updateField = (index, field) => {
    const newFields = [...formData.fields];
    newFields[index] = { ...newFields[index], ...field };
    setFormData({ ...formData, fields: newFields });
  };

  const removeField = (index) => {
    const newFields = formData.fields.filter((_, i) => i !== index);
    setFormData({ ...formData, fields: newFields });
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    
    // Validate
    if (!formData.name.trim()) {
      showAlert('Form name is required', 'error');
      return;
    }
    if (!formData.user_type) {
      showAlert('User type is required', 'error');
      return;
    }

    // Filter out empty fields
    const validFields = formData.fields.filter(f => f.name && f.label);

    try {
      const templateData = {
        name: formData.name.trim(),
        description: formData.description.trim(),
        user_type: formData.user_type,
        fields: validFields
      };

      if (editingForm) {
        await updateFormTemplate(editingForm, templateData);
        showAlert('Form updated successfully!', 'success');
      } else {
        await createFormTemplate(templateData);
        showAlert('Form created successfully!', 'success');
      }
      
      closeModal();
      loadForms();
    } catch (error) {
      showAlert('Error saving form: ' + (error.response?.data?.error || error.message), 'error');
    }
  };

  const handleDelete = async (id) => {
    if (!window.confirm('Are you sure you want to delete this form?')) return;
    
    try {
      await deleteFormTemplate(id);
      showAlert('Form deleted successfully!', 'success');
      loadForms();
    } catch (error) {
      showAlert('Error deleting form: ' + (error.response?.data?.error || error.message), 'error');
    }
  };

  const formatDate = (dateString) => {
    if (!dateString) return 'N/A';
    try {
      return new Date(dateString).toLocaleDateString();
    } catch {
      return dateString;
    }
  };

  return (
    <div className="forms-container">
      <div className="forms-header">
        <h1>Form Management</h1>
        <p>Create and manage form templates for students and staff</p>
      </div>

      <div className="forms-content">
        {alert && (
          <div className={`alert alert-${alert.type}`}>
            {alert.message}
          </div>
        )}

        <div className="toolbar">
          <div className="filter-group">
            <label htmlFor="userTypeFilter">Filter:</label>
            <select
              id="userTypeFilter"
              value={userTypeFilter}
              onChange={(e) => setUserTypeFilter(e.target.value)}
            >
              <option value="">All Types</option>
              <option value="student">Student Forms</option>
              <option value="staff">Staff Forms</option>
            </select>
          </div>
          <button className="btn" onClick={openCreateModal}>
            + Create New Form
          </button>
        </div>

        {loading ? (
          <div className="loading">Loading forms...</div>
        ) : (filteredForms || []).length === 0 ? (
          <div className="empty-state">
            <h3>No forms found</h3>
            <p>Create your first form template to get started</p>
          </div>
        ) : (
          <div className="forms-grid">
            {(filteredForms || []).map(form => (
              <div key={form.id} className="form-card">
                <h3>{form.name}</h3>
                <div className="meta">
                  <strong>Type:</strong> {form.user_type} | 
                  <strong> Fields:</strong> {form.fields?.length || 0} | 
                  <strong> Created:</strong> {formatDate(form.created_at)}
                </div>
                {form.description && (
                  <p style={{ color: '#b0b0b0', margin: '0.5rem 0' }}>{form.description}</p>
                )}
                <div className="fields-preview">
                  {form.fields?.map((f, idx) => (
                    <span key={idx} className="field-tag">
                      {f.label || f.name}
                    </span>
                  )) || <span className="field-tag">No fields</span>}
                </div>
                <div className="actions">
                  <button
                    className="btn btn-secondary btn-small"
                    onClick={() => openEditModal(form.id)}
                  >
                    Edit
                  </button>
                  <button
                    className="btn btn-danger btn-small"
                    onClick={() => handleDelete(form.id)}
                  >
                    Delete
                  </button>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>

      {/* Modal */}
      {showModal && (
        <div className="modal-overlay" onClick={closeModal}>
          <div className="modal-content" onClick={(e) => e.stopPropagation()}>
            <div className="modal-header">
              <h2>{editingForm ? 'Edit Form' : 'Create New Form'}</h2>
              <button className="close-btn" onClick={closeModal}>&times;</button>
            </div>
            <form onSubmit={handleSubmit}>
              <div className="form-group">
                <label htmlFor="formName">Form Name *</label>
                <input
                  type="text"
                  id="formName"
                  value={formData.name}
                  onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                  required
                />
              </div>
              <div className="form-group">
                <label htmlFor="formDescription">Description</label>
                <textarea
                  id="formDescription"
                  value={formData.description}
                  onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                />
              </div>
              <div className="form-group">
                <label htmlFor="formUserType">User Type *</label>
                <select
                  id="formUserType"
                  value={formData.user_type}
                  onChange={(e) => setFormData({ ...formData, user_type: e.target.value })}
                  required
                >
                  <option value="">Select...</option>
                  <option value="student">Student</option>
                  <option value="staff">Staff</option>
                </select>
              </div>
              <div className="form-group">
                <label>Form Fields</label>
                <div className="fields-editor">
                  {formData.fields.map((field, index) => (
                    <div key={index} className="field-item">
                      <div className="field-item-header">
                        <h4>Field {index + 1}</h4>
                        <button
                          type="button"
                          className="btn btn-danger btn-small"
                          onClick={() => removeField(index)}
                        >
                          Remove
                        </button>
                      </div>
                      <div className="field-row">
                        <div className="form-group">
                          <label>Field Name (ID) *</label>
                          <input
                            type="text"
                            value={field.name}
                            onChange={(e) => updateField(index, { name: e.target.value })}
                            placeholder="e.g., name, age"
                            required
                          />
                        </div>
                        <div className="form-group">
                          <label>Label *</label>
                          <input
                            type="text"
                            value={field.label}
                            onChange={(e) => updateField(index, { label: e.target.value })}
                            placeholder="e.g., Full Name"
                            required
                          />
                        </div>
                      </div>
                      <div className="field-row">
                        <div className="form-group">
                          <label>Type *</label>
                          <select
                            value={field.type}
                            onChange={(e) => updateField(index, { type: e.target.value })}
                            required
                          >
                            <option value="text">Text</option>
                            <option value="email">Email</option>
                            <option value="number">Number</option>
                            <option value="tel">Phone</option>
                            <option value="date">Date</option>
                            <option value="select">Select</option>
                          </select>
                        </div>
                        <div className="form-group">
                          <label>Placeholder</label>
                          <input
                            type="text"
                            value={field.placeholder || ''}
                            onChange={(e) => updateField(index, { placeholder: e.target.value })}
                            placeholder="Placeholder text"
                          />
                        </div>
                      </div>
                      <div className="field-row">
                        <div className="form-group checkbox-group">
                          <input
                            type="checkbox"
                            checked={field.required}
                            onChange={(e) => updateField(index, { required: e.target.checked })}
                          />
                          <label>Required Field</label>
                        </div>
                      </div>
                      {field.type === 'select' && (
                        <div className="field-row full">
                          <div className="form-group">
                            <label>Options (comma-separated)</label>
                            <input
                              type="text"
                              value={field.options ? field.options.join(', ') : ''}
                              onChange={(e) => {
                                const options = e.target.value
                                  .split(',')
                                  .map(o => o.trim())
                                  .filter(o => o);
                                updateField(index, { options });
                              }}
                              placeholder="Option 1, Option 2, Option 3"
                            />
                          </div>
                        </div>
                      )}
                    </div>
                  ))}
                  <button
                    type="button"
                    className="btn btn-secondary btn-small"
                    onClick={addField}
                  >
                    + Add Field
                  </button>
                </div>
              </div>
              <div className="modal-actions">
                <button type="button" className="btn btn-secondary" onClick={closeModal}>
                  Cancel
                </button>
                <button type="submit" className="btn">
                  Save Form
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
};

export default Forms;
