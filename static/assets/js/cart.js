// assets/js/carts.js

/**
 * @param {number|string} amount - Jumlah yang akan diformat.
 * @returns {string} - String format Rupiah.
 */
function formatCurrency(amount) {
    const num = parseFloat(amount);
    if (isNaN(num)) {
        return "Rp 0";
    }
    return new Intl.NumberFormat('id-ID', {
        style: 'currency',
        currency: 'IDR',
        minimumFractionDigits: 0
    }).format(num);
}

document.addEventListener('DOMContentLoaded', function() {

    const flashMessage = document.getElementById('flash-message');
    if (flashMessage) {
        setTimeout(() => {
            flashMessage.classList.add('animate-fade-out');
            flashMessage.addEventListener('animationend', () => {
                flashMessage.remove();
            });
        }, 5000);
    }

    // Ambil elemen dengan aman, default ke 0 jika null
    const cartGrandTotalAmountInput = document.getElementById('cart_grand_total_amount_for_js');
    const totalWeightInput = document.getElementById('total_weight_input');
    const originLocationIDInput = document.getElementById('origin_location_id_input');

    const elements = {
        addressSelect: document.getElementById('address_id'),
        courierSelect: document.getElementById('courier_select'),
        shippingFeeSelect: document.getElementById('shipping_fee_options'),
        shippingCalculationMsg: document.getElementById('shipping-calculation-msg'),
        shippingFeeDisplay: document.getElementById('shipping-fee-display'),
        grandTotalDisplay: document.getElementById('grand-total-display'),
        // Safely get values, default to 0 or 1 if element is null
        cartGrandTotalBeforeShipping: cartGrandTotalAmountInput ? parseFloat(cartGrandTotalAmountInput.value) : 0,
        totalWeightInput: totalWeightInput ? parseFloat(totalWeightInput.value) : 1, // Default ke 1 jika null atau 0
        proceedToCheckoutBtn: document.getElementById('proceedToCheckoutBtn'),

        checkoutShippingCostInput: document.getElementById('checkout_shipping_cost'),
        checkoutShippingServiceCodeInput: document.getElementById('checkout_shipping_service_code'),
        checkoutShippingServiceNameInput: document.getElementById('checkout_shipping_service_name'),
        checkoutFinalTotalPriceInput: document.getElementById('checkout_final_total_price'),
    };

    let selectedDestinationLocationID = null;
    let totalWeight = elements.totalWeightInput;
    let cartGrandTotal = elements.cartGrandTotalBeforeShipping;

    // Ensure totalWeight is at least 1 gram if 0
    if (totalWeight <= 0) {
        totalWeight = 1;
    }

    // Ambil originLocationID dari hidden input
    const originLocationID = originLocationIDInput ? parseInt(originLocationIDInput.value, 10) : 0;
    if (originLocationID === 0) {
        console.error('Error: Origin Location ID tidak ditemukan atau tidak valid dari backend.');
        if (elements.shippingCalculationMsg) { // Pastikan elemen ada sebelum diakses
            setShippingMessage('Konfigurasi toko asal pengiriman tidak valid. Mohon hubungi admin.', 'error');
        }
    }


    /**
     * Updates the displayed grand total.
     * @param {number|string} selectedShippingCostValue - The selected shipping cost.
     */
    function updateGrandTotalDisplay(selectedShippingCostValue = 0) {
        const currentShippingFee = parseFloat(selectedShippingCostValue) || 0;
        const grandTotalCalculated = cartGrandTotal + currentShippingFee;

        if (elements.shippingFeeDisplay) elements.shippingFeeDisplay.textContent = formatCurrency(currentShippingFee);
        if (elements.grandTotalDisplay) elements.grandTotalDisplay.textContent = formatCurrency(grandTotalCalculated);
        if (elements.checkoutFinalTotalPriceInput) elements.checkoutFinalTotalPriceInput.value = grandTotalCalculated.toFixed(2); // Update hidden input
    }

    /**
     * Resets shipping related elements (except courier select disabled state).
     */
    function resetShippingOptionsAndButton() {
        if (elements.shippingFeeSelect) { // PERBAIKAN: Cek null
            elements.shippingFeeSelect.innerHTML = `<option value="" selected>--Pilih Opsi Pengiriman--</option>`;
            elements.shippingFeeSelect.disabled = true;
        }
        
        if (elements.checkoutShippingCostInput) elements.checkoutShippingCostInput.value = 0;
        if (elements.checkoutShippingServiceCodeInput) elements.checkoutShippingServiceCodeInput.value = '';
        if (elements.checkoutShippingServiceNameInput) elements.checkoutShippingServiceNameInput.value = '';
        
        if (elements.proceedToCheckoutBtn) elements.proceedToCheckoutBtn.disabled = true;
        setShippingMessage('');
        updateGrandTotalDisplay(0); // Reset total payment to initial cart total
    }

    /**
     * Displays a message related to shipping calculation.
     * @param {string} message - The message to display.
     * @param {'success'|'warning'|'error'|''} type - Message type (for styling).
     */
    function setShippingMessage(message, type = '') {
        if (!elements.shippingCalculationMsg) { // PERBAIKAN: Cek null
            console.error('Error: shippingCalculationMsg element is null. Cannot display message:', message);
            return;
        }
        elements.shippingCalculationMsg.textContent = message;
        elements.shippingCalculationMsg.className = 'text-sm mt-2'; // Reset classes

        elements.shippingCalculationMsg.classList.remove('text-red-500', 'text-yellow-500', 'text-green-500');
        if (type === 'error') {
            elements.shippingCalculationMsg.classList.add('text-red-500');
        } else if (type === 'warning') {
            elements.shippingCalculationMsg.classList.add('text-yellow-500');
        } else if (type === 'success') {
            elements.shippingCalculationMsg.classList.add('text-green-500');
        }
    }

    // Function to calculate shipping cost using Komerce API
    async function calculateShippingCost() {
        const destinationID = selectedDestinationLocationID;
        const courier = elements.courierSelect.value;

        resetShippingOptionsAndButton(); // Only reset shipping options and button here

        console.log('--- Calculating Shipping Cost ---');
        console.log('originLocationID (from JS):', originLocationID);
        console.log('destinationLocationID (from JS):', destinationID);
        console.log('totalWeight:', totalWeight);
        console.log('courier:', courier);

        if (originLocationID === 0) {
            setShippingMessage('Kota asal pengiriman belum ditentukan (konfigurasi backend).', 'error');
            return;
        }
        if (totalWeight <= 0) {
            setShippingMessage('Berat total produk harus lebih dari 0.', 'error');
            return;
        }
        if (!destinationID || parseInt(destinationID, 10) === 0) {
            setShippingMessage('Mohon pilih Alamat Pengiriman yang valid.', 'warning');
            return;
        }
        if (!courier) {
            setShippingMessage('Mohon pilih Kurir.', 'warning');
            return;
        }

        setShippingMessage('Memuat opsi pengiriman...', 'warning');
        if (elements.shippingFeeSelect) { // Pastikan elemen ada sebelum diakses
            elements.shippingFeeSelect.innerHTML = '<option value="" selected>--Memuat Opsi Pengiriman--</option>';
        }
        
        // Dapatkan CSRF token dari hidden input di form checkout
        const checkoutForm = document.getElementById('checkout-form');
        const csrfTokenInput = checkoutForm ? checkoutForm.querySelector('input[name="csrf_token"]') : null;
        const csrfToken = csrfTokenInput ? csrfTokenInput.value : '';

        if (!csrfToken) {
            console.error('CSRF token not found in checkout form.');
            setShippingMessage('Kesalahan: CSRF token tidak ditemukan. Mohon refresh halaman.', 'error');
            return;
        }

        try {
            const response = await fetch('/api/komerce/calculate-shipping-cost', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json', // Sending JSON
                    'X-CSRF-Token': csrfToken, // KIRIM CSRF TOKEN DI HEADER INI
                },
                body: JSON.stringify({
                    origin: originLocationID,
                    destination: parseInt(destinationID, 10), // Ensure this is an int
                    weight: parseInt(totalWeight, 10), // Ensure this is an int
                    courier: courier,
                })
            });

            // Periksa apakah respons adalah HTML (bukan JSON)
            const contentType = response.headers.get("content-type");
            if (!contentType || !contentType.includes("application/json")) {
                const errorText = await response.text();
                console.error("Received non-JSON response:", errorText);
                throw new Error(`Server returned non-JSON response. Status: ${response.status}. Content: ${errorText.substring(0, 100)}...`);
            }

            const data = await response.json();

            if (!response.ok || data.status === 'error' || !data.success) { // Check data.success as well
                let displayMessage = data.message || `Error ${response.status}: Gagal menghitung ongkos kirim.`;
                
                if (response.status === 404 && data.meta && data.meta.message && data.meta.message.includes("Calculate Domestic Shipping Cost not found")) {
                    displayMessage = `Opsi pengiriman tidak tersedia untuk kurir '${courier}' pada rute ini. Mohon coba kurir lain atau hubungi dukungan.`;
                } else if (response.status === 400 && data.meta && data.meta.message) {
                    displayMessage = `Kesalahan validasi dari API: ${data.meta.message}. Mohon periksa input Anda.`;
                }

                throw new Error(displayMessage);
            }

            if (elements.shippingFeeSelect) { // Pastikan elemen ada sebelum diakses
                elements.shippingFeeSelect.innerHTML = '<option value="" selected>--Pilih Opsi Pengiriman--</option>';
            }

            let optionsFound = false;
            if (data.data && data.data.length > 0) {
                data.data.forEach(service => {
                    const serviceName = `${service.name} - ${service.service} (${service.description})`;
                    const costValue = service.cost;
                    const etd = service.etd;

                    const optionText = `${serviceName} - ${formatCurrency(costValue)} (Estimasi: ${etd} hari)`;
                    const optionValue = `${costValue}|${service.code}|${service.service}|${serviceName}`;
                    const option = new Option(optionText, optionValue);
                    if (elements.shippingFeeSelect) { // Pastikan elemen ada sebelum diakses
                        elements.shippingFeeSelect.add(option);
                    }
                    optionsFound = true;
                });
            }

            if (optionsFound) {
                if (elements.shippingFeeSelect) elements.shippingFeeSelect.disabled = false;
                setShippingMessage('Pilih opsi pengiriman.', 'success');
            } else {
                setShippingMessage('Tidak ada opsi pengiriman tersedia untuk rute dan kurir ini. Coba kurir lain atau periksa rute.', 'warning');
                if (elements.shippingFeeSelect) elements.shippingFeeSelect.innerHTML = '<option value="" selected>--Tidak Tersedia--</option>';
            }
        } catch (error) {
            console.error('Error dalam calculateShippingCost:', error);
            setShippingMessage(error.message || 'Terjadi kesalahan jaringan saat menghitung ongkir.', 'error');
            if (elements.shippingFeeSelect) elements.shippingFeeSelect.innerHTML = '<option value="" selected>--Gagal Memuat--</option>';
        } finally {
            updateGrandTotalDisplay(parseFloat(elements.checkoutShippingCostInput.value));
        }
    }

    // --- Event Listeners ---

    // Event listener for address selection change
    if (elements.addressSelect) { // PERBAIKAN: Tambahkan null check
        elements.addressSelect.addEventListener('change', function() {
            const selectedOption = this.options[this.selectedIndex];
            if (selectedOption.value) {
                selectedDestinationLocationID = selectedOption.dataset.locationId;
                if (elements.courierSelect) elements.courierSelect.disabled = false; // AKTIFKAN KURIR DI SINI
                if (elements.courierSelect) elements.courierSelect.value = ""; // Reset courier selection
                resetShippingOptionsAndButton(); // Hanya reset opsi pengiriman dan tombol
                setShippingMessage('Pilih kurir untuk menghitung ongkos kirim.', 'success');
            } else {
                selectedDestinationLocationID = null;
                if (elements.courierSelect) elements.courierSelect.value = ""; // Reset courier selection
                if (elements.courierSelect) elements.courierSelect.disabled = true; // NONAKTIFKAN KURIR JIKA TIDAK ADA ALAMAT
                resetShippingOptionsAndButton(); // Reset opsi pengiriman dan tombol
                setShippingMessage('Mohon pilih alamat pengiriman.', 'warning');
            }
        });
    }

    // Event listener for courier selection change
    if (elements.courierSelect) { // PERBAIKAN: Tambahkan null check
        elements.courierSelect.addEventListener('change', function() {
            if (this.value && selectedDestinationLocationID) {
                calculateShippingCost();
            } else {
                resetShippingOptionsAndButton(); // Hanya reset opsi pengiriman dan tombol
                setShippingMessage('Pilih kurir untuk menghitung ongkos kirim.', 'warning');
            }
        });
    }

    // Event listener for shipping option selection change
    if (elements.shippingFeeSelect) { // PERBAIKAN: Tambahkan null check
        elements.shippingFeeSelect.addEventListener('change', function() {
            const selectedOptionValue = this.value;
            if (selectedOptionValue) {
                const parts = selectedOptionValue.split('|');
                const cost = parseFloat(parts[0]);
                const serviceCode = parts[1];
                const serviceName = parts[3];

                if (elements.checkoutShippingCostInput) elements.checkoutShippingCostInput.value = cost.toFixed(2);
                if (elements.checkoutShippingServiceCodeInput) elements.checkoutShippingServiceCodeInput.value = serviceCode;
                if (elements.checkoutShippingServiceNameInput) elements.checkoutShippingServiceNameInput.value = serviceName;

                updateGrandTotalDisplay(cost);
                if (elements.proceedToCheckoutBtn) elements.proceedToCheckoutBtn.disabled = false;
            } else {
                if (elements.checkoutShippingCostInput) elements.checkoutShippingCostInput.value = 0;
                if (elements.checkoutShippingServiceCodeInput) elements.checkoutShippingServiceCodeInput.value = '';
                if (elements.checkoutShippingServiceNameInput) elements.checkoutShippingServiceNameInput.value = '';
                updateGrandTotalDisplay(0);
                if (elements.proceedToCheckoutBtn) elements.proceedToCheckoutBtn.disabled = true;
            }
        });
    }

    // SweetAlert for delete cart item confirmation
    document.querySelectorAll('.delete-cart-item-form').forEach(form => {
        form.addEventListener('submit', function(e) {
            e.preventDefault();
            const formElement = this;

            Swal.fire({
                title: 'Apakah Anda yakin?',
                text: "Item ini akan dihapus dari keranjang!",
                icon: 'warning',
                showCancelButton: true,
                confirmButtonColor: '#d33',
                cancelButtonColor: '#3085d6',
                confirmButtonText: 'Ya, hapus!',
                cancelButtonText: 'Batal'
            }).then((result) => {
                if (result.isConfirmed) {
                    formElement.submit();
                }
            });
        });
    });

    // SweetAlert for update cart item confirmation
    document.querySelectorAll('.update-cart-item-form').forEach(form => {
        form.addEventListener('submit', function(e) {
            e.preventDefault();
            const formElement = this;

            Swal.fire({
                title: 'Perbarui Jumlah?',
                text: "Jumlah item di keranjang akan diperbarui.",
                icon: 'question',
                showCancelButton: true,
                confirmButtonColor: '#28a745',
                cancelButtonColor: '#dc3545',
                confirmButtonText: 'Ya, perbarui!',
                cancelButtonText: 'Batal'
            }).then((result) => {
                if (result.isConfirmed) {
                    formElement.submit();
                }
            });
        });
    });

    // Inisialisasi awal jika ada alamat yang terpilih secara default atau dari sesi
    if (elements.addressSelect && elements.addressSelect.value) {
        const initialSelectedOption = elements.addressSelect.options[elements.addressSelect.selectedIndex];
        selectedDestinationLocationID = initialSelectedOption.dataset.locationId;
        if (elements.courierSelect) elements.courierSelect.disabled = false;
        setShippingMessage('Pilih kurir untuk menghitung ongkos kirim.', 'success');
    } else {
        setShippingMessage('Mohon pilih alamat pengiriman.', 'warning');
    }

}); // End DOMContentLoaded
