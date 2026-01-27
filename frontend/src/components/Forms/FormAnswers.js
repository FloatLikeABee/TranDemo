import React, { useState, useEffect } from 'react';
import {
  getFormAnswers,
  getFormAnswer,
  createFormAnswer,
  updateFormAnswer,
  deleteFormAnswer,
  getFormTemplates,
  getFormTemplate
} from '../../services/formsApi';
import './FormAnswers.css';
import './Forms.css';

const FormAnswers = () => {
  const [answers, setAnswers] = useState([]);
  const [filteredAnswers, setFilteredAnswers] = useState([]);
  const [forms, setForms] = useState([]);
  const [loading, setLoading] = useState(true);
  const [viewMode, setViewMode] = useState('grid');
  const [formFilter, setFormFilter] = useState('');
  const [userTypeFilter, setUserTypeFilter] = useState('');
  const [userIdFilter, setUserIdFilter] = useState('');
  const [showModal, setShowModal] = useState(false);
  const [editingAnswer, setEditingAnswer] = useState(null);
  const [alert, setAlert] = useState(null);
  const [answerData, setAnswerData] = useState({
    form_id: '',
    user_id: '',
    user_type: '',
    answers: {}
  });
  const [selectedFormFields, setSelectedFormFields] = useState([]);

  useEffect(() => {
    loadForms();
    loadAnswers();
  }, []);

  useEffect(() => {
    filterAnswers();
  }, [answers, formFilter, userTypeFilter, userIdFilter]);

  const loadForms = async () => {
    try {
      const data = await getFormTemplates();
      setForms(data);
    } catch (error) {
      showAlert('Error loading forms: ' + (error.response?.data?.error || error.message), 'error');
    }
  };

  const loadAnswers = async () => {
    try {
      setLoading(true);
      const data = await getFormAnswers();
      setAnswers(data);
    } catch (error) {
      showAlert('Error loading answers: ' + (error.response?.data?.error || error.message), 'error');
    } finally {
      setLoading(false);
    }
  };

  const filterAnswers = () => {
    let filtered = answers;
    
    if (formFilter) {
      filtered = filtered.filter(a => a.form_id === formFilter);
    }
    if (userTypeFilter) {
      filtered = filtered.filter(a => a.user_type === userTypeFilter);
    }
    if (userIdFilter) {
      filtered = filtered.filter(a => 
        a.user_id.toLowerCase().includes(userIdFilter.toLowerCase())
      );
    }
    
    setFilteredAnswers(filtered);
  };

  const showAlert = (message, type) => {
    setAlert({ message, type });
    setTimeout(() => setAlert(null), 5000);
  };

  const loadFormFields = async (formId) => {
    if (!formId) {
      setSelectedFormFields([]);
      return;
    }

    try {
      const form = await getFormTemplate(formId);
      setSelectedFormFields(form.fields || []);
      
      // Initialize answer data with empty values for each field
      const initialAnswers = {};
      form.fields.forEach(field => {
        initialAnswers[field.name] = '';
      });
      setAnswerData(prev => ({
        ...prev,
        answers: { ...prev.answers, ...initialAnswers }
      }));
    } catch (error) {
      showAlert('Error loading form fields: ' + (error.response?.data?.error || error.message), 'error');
    }
  };

  const openCreateModal = () => {
    setEditingAnswer(null);
    setAnswerData({
      form_id: '',
      user_id: '',
      user_type: '',
      answers: {}
    });
    setSelectedFormFields([]);
    setShowModal(true);
  };

  const openEditModal = async (id) => {
    try {
      const answer = await getFormAnswer(id);
      setEditingAnswer(id);
      setAnswerData({
        form_id: answer.form_id,
        user_id: answer.user_id,
        user_type: answer.user_type,
        answers: answer.answers || {}
      });
      
      // Load form fields
      await loadFormFields(answer.form_id);
      setShowModal(true);
    } catch (error) {
      showAlert('Error loading answer: ' + (error.response?.data?.error || error.message), 'error');
    }
  };

  const closeModal = () => {
    setShowModal(false);
    setEditingAnswer(null);
    setAnswerData({
      form_id: '',
      user_id: '',
      user_type: '',
      answers: {}
    });
    setSelectedFormFields([]);
  };

  const handleFormChange = async (formId) => {
    setAnswerData(prev => ({ ...prev, form_id: formId, answers: {} }));
    await loadFormFields(formId);
  };

  const handleAnswerChange = (fieldName, value) => {
    setAnswerData(prev => ({
      ...prev,
      answers: {
        ...prev.answers,
        [fieldName]: value
      }
    }));
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    
    // Validate
    if (!answerData.form_id) {
      showAlert('Form template is required', 'error');
      return;
    }
    if (!answerData.user_id.trim()) {
      showAlert('User ID is required', 'error');
      return;
    }
    if (!answerData.user_type) {
      showAlert('User type is required', 'error');
      return;
    }

    try {
      if (editingAnswer) {
        await updateFormAnswer(editingAnswer, answerData);
        showAlert('Answer updated successfully!', 'success');
      } else {
        await createFormAnswer(answerData);
        showAlert('Answer submitted successfully!', 'success');
      }
      
      closeModal();
      loadAnswers();
    } catch (error) {
      showAlert('Error saving answer: ' + (error.response?.data?.error || error.message), 'error');
    }
  };

  const handleDelete = async (id) => {
    if (!window.confirm('Are you sure you want to delete this answer?')) return;
    
    try {
      await deleteFormAnswer(id);
      showAlert('Answer deleted successfully!', 'success');
      loadAnswers();
    } catch (error) {
      showAlert('Error deleting answer: ' + (error.response?.data?.error || error.message), 'error');
    }
  };

  const formatDate = (dateString) => {
    if (!dateString) return 'N/A';
    try {
      return new Date(dateString).toLocaleString();
    } catch {
      return dateString;
    }
  };

  return (
    <div className="form-answers-container">
      <div className="form-answers-header">
        <h1>Form Answers</h1>
        <p>View and manage submitted form answers</p>
      </div>

      <div className="form-answers-content">
        {alert && (
          <div className={`alert alert-${alert.type}`}>
            {alert.message}
          </div>
        )}

        <div className="toolbar">
          <div className="filter-group">
            <label htmlFor="formFilter">Form:</label>
            <select
              id="formFilter"
              value={formFilter}
              onChange={(e) => setFormFilter(e.target.value)}
            >
              <option value="">All Forms</option>
              {forms.map(form => (
                <option key={form.id} value={form.id}>{form.name}</option>
              ))}
            </select>
            <label htmlFor="userTypeFilter">User Type:</label>
            <select
              id="userTypeFilter"
              value={userTypeFilter}
              onChange={(e) => setUserTypeFilter(e.target.value)}
            >
              <option value="">All Types</option>
              <option value="student">Student</option>
              <option value="staff">Staff</option>
            </select>
            <label htmlFor="userIdFilter">User ID:</label>
            <input
              type="text"
              id="userIdFilter"
              value={userIdFilter}
              onChange={(e) => setUserIdFilter(e.target.value)}
              placeholder="Filter by user ID"
              style={{
                padding: '0.5rem 1rem',
                background: '#1a1a1a',
                color: '#e0e0e0',
                border: '1px solid #3a3a3a',
                borderRadius: '4px',
                fontSize: '1rem'
              }}
            />
          </div>
          <button className="btn" onClick={openCreateModal}>
            + Submit New Answer
          </button>
        </div>

        <div className="view-toggle">
          <button
            className={viewMode === 'grid' ? 'active' : ''}
            onClick={() => setViewMode('grid')}
          >
            Grid View
          </button>
          <button
            className={viewMode === 'table' ? 'active' : ''}
            onClick={() => setViewMode('table')}
          >
            Table View
          </button>
        </div>

        {loading ? (
          <div className="loading">Loading answers...</div>
        ) : filteredAnswers.length === 0 ? (
          <div className="empty-state">
            <h3>No answers found</h3>
            <p>Submit your first form answer to get started</p>
          </div>
        ) : viewMode === 'grid' ? (
          <div className="answers-grid">
            {filteredAnswers.map(answer => (
              <div key={answer.id} className="answer-card">
                <h3>{answer.form_name || 'Unknown Form'}</h3>
                <div className="meta">
                  <span className={`badge badge-${answer.user_type}`}>
                    {answer.user_type}
                  </span>
                  <strong> User ID:</strong> {answer.user_id} | 
                  <strong> Submitted:</strong> {formatDate(answer.submitted_at)}
                </div>
                <div className="answers-list">
                  {Object.entries(answer.answers || {}).map(([key, value]) => (
                    <div key={key} className="answer-item">
                      <strong>{key}</strong>
                      <span>{String(value)}</span>
                    </div>
                  ))}
                </div>
                <div style={{ display: 'flex', gap: '0.5rem', marginTop: '1rem' }}>
                  <button
                    className="btn btn-secondary btn-small"
                    onClick={() => openEditModal(answer.id)}
                  >
                    Edit
                  </button>
                  <button
                    className="btn btn-danger btn-small"
                    onClick={() => handleDelete(answer.id)}
                  >
                    Delete
                  </button>
                </div>
              </div>
            ))}
          </div>
        ) : (
          <table className="answers-table">
            <thead>
              <tr>
                <th>Form Name</th>
                <th>User ID</th>
                <th>User Type</th>
                <th>Submitted At</th>
                <th>Actions</th>
              </tr>
            </thead>
            <tbody>
              {filteredAnswers.map(answer => (
                <tr key={answer.id}>
                  <td>{answer.form_name || 'Unknown'}</td>
                  <td>{answer.user_id}</td>
                  <td>
                    <span className={`badge badge-${answer.user_type}`}>
                      {answer.user_type}
                    </span>
                  </td>
                  <td>{formatDate(answer.submitted_at)}</td>
                  <td>
                    <button
                      className="btn btn-secondary btn-small"
                      onClick={() => openEditModal(answer.id)}
                    >
                      Edit
                    </button>
                    <button
                      className="btn btn-danger btn-small"
                      onClick={() => handleDelete(answer.id)}
                    >
                      Delete
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>

      {/* Modal */}
      {showModal && (
        <div className="modal-overlay" onClick={closeModal}>
          <div className="modal-content" onClick={(e) => e.stopPropagation()}>
            <div className="modal-header">
              <h2>{editingAnswer ? 'Edit Answer' : 'Submit New Answer'}</h2>
              <button className="close-btn" onClick={closeModal}>&times;</button>
            </div>
            <form onSubmit={handleSubmit}>
              <div className="form-group">
                <label htmlFor="answerFormID">Form Template *</label>
                <select
                  id="answerFormID"
                  value={answerData.form_id}
                  onChange={(e) => handleFormChange(e.target.value)}
                  required
                >
                  <option value="">Select a form...</option>
                  {forms.map(form => (
                    <option key={form.id} value={form.id}>{form.name}</option>
                  ))}
                </select>
              </div>
              <div className="form-group">
                <label htmlFor="answerUserID">User ID *</label>
                <input
                  type="text"
                  id="answerUserID"
                  value={answerData.user_id}
                  onChange={(e) => setAnswerData({ ...answerData, user_id: e.target.value })}
                  required
                />
              </div>
              <div className="form-group">
                <label htmlFor="answerUserType">User Type *</label>
                <select
                  id="answerUserType"
                  value={answerData.user_type}
                  onChange={(e) => setAnswerData({ ...answerData, user_type: e.target.value })}
                  required
                >
                  <option value="">Select...</option>
                  <option value="student">Student</option>
                  <option value="staff">Staff</option>
                </select>
              </div>
              <div className="form-group">
                <label>Answers</label>
                <div className="answers-editor">
                  {selectedFormFields.length === 0 ? (
                    <p style={{ color: '#b0b0b0' }}>Select a form template to load fields</p>
                  ) : (
                    selectedFormFields.map((field, index) => (
                      <div key={index} className="answer-field-item">
                        <label>
                          {field.label} {field.required && <span style={{ color: '#dc3545' }}>*</span>}
                        </label>
                        {field.type === 'select' && field.options && field.options.length > 0 ? (
                          <select
                            value={answerData.answers[field.name] || ''}
                            onChange={(e) => handleAnswerChange(field.name, e.target.value)}
                            required={field.required}
                          >
                            <option value="">Select...</option>
                            {field.options.map((opt, idx) => (
                              <option key={idx} value={opt}>{opt}</option>
                            ))}
                          </select>
                        ) : (
                          <input
                            type={field.type === 'number' ? 'number' : 
                                  field.type === 'email' ? 'email' : 
                                  field.type === 'tel' ? 'tel' : 
                                  field.type === 'date' ? 'date' : 'text'}
                            value={answerData.answers[field.name] || ''}
                            onChange={(e) => handleAnswerChange(field.name, e.target.value)}
                            placeholder={field.placeholder || ''}
                            required={field.required}
                          />
                        )}
                      </div>
                    ))
                  )}
                </div>
              </div>
              <div className="modal-actions">
                <button type="button" className="btn btn-secondary" onClick={closeModal}>
                  Cancel
                </button>
                <button type="submit" className="btn">
                  Save Answer
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
};

export default FormAnswers;
