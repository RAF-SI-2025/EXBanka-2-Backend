DELETE FROM core_banking.exchange
WHERE mic_code IN (
    'XNYS','XNAS','XASE','XCBO',
    'XLON',
    'XETR','XPAR','XMIL','XAMS','XBRU','XMAD','XVIE','XLIS','XDUB','ASEX','XHEL',
    'XTKS','XOSE',
    'XTSE','XTSX',
    'XASX',
    'XSWX',
    'XBEL'
);
