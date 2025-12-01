// NeonEx Framework - Main JavaScript

document.addEventListener('DOMContentLoaded', function() {
    console.log('NeonEx Framework initialized');
    
    // Mobile menu toggle
    const menuToggle = document.querySelector('.menu-toggle');
    const navLinks = document.querySelector('.nav-links');
    
    if (menuToggle) {
        menuToggle.addEventListener('click', function() {
            navLinks.classList.toggle('active');
        });
    }
    
    // Smooth scroll
    document.querySelectorAll('a[href^="#"]').forEach(anchor => {
        anchor.addEventListener('click', function (e) {
            e.preventDefault();
            const target = document.querySelector(this.getAttribute('href'));
            if (target) {
                target.scrollIntoView({
                    behavior: 'smooth'
                });
            }
        });
    });
});

// Utility functions
const NeonEx = {
    // Show notification
    notify: function(message, type = 'info') {
        console.log(`[${type.toUpperCase()}] ${message}`);
        // TODO: Implement toast notification
    },
    
    // AJAX helper
    ajax: async function(url, options = {}) {
        try {
            const response = await fetch(url, {
                ...options,
                headers: {
                    'Content-Type': 'application/json',
                    ...options.headers
                }
            });
            return await response.json();
        } catch (error) {
            console.error('AJAX Error:', error);
            throw error;
        }
    },
    
    // Format currency
    formatCurrency: function(amount, currency = 'THB') {
        return new Intl.NumberFormat('th-TH', {
            style: 'currency',
            currency: currency
        }).format(amount);
    },
    
    // Format date
    formatDate: function(date) {
        return new Intl.DateTimeFormat('th-TH').format(new Date(date));
    }
};

// Export to global scope
window.NeonEx = NeonEx;
