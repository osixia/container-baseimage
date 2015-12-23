# to geneate a new ca :
cfssl gencert -initca config/ca-csr.json | cfssljson -bare default-ca
