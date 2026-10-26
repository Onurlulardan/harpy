SELECT side,kind,identity,document
FROM catalog_adoption_inventory_v1($1::uuid)
ORDER BY side,kind,identity
