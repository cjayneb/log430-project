document.addEventListener("DOMContentLoaded", () => {
    const controlField = document.getElementById('type');
    const dependentField = document.getElementById('unit_price');

    function updateDependentFieldRequirement() {
        if (controlField.value === 'limit') {
            dependentField.removeAttribute('disabled');
            dependentField.setAttribute('required', true);
        } else {
            dependentField.removeAttribute('required');
            dependentField.setAttribute('disabled', true);
        }
    }

    controlField.addEventListener('change', updateDependentFieldRequirement);
    updateDependentFieldRequirement();
});