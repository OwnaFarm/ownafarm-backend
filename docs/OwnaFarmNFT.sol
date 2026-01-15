// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import {ERC1155} from "@openzeppelin/contracts/token/ERC1155/ERC1155.sol";
import {AccessControl} from "@openzeppelin/contracts/access/AccessControl.sol";
import {IERC20} from "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import {SafeERC20} from "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";
import {ReentrancyGuard} from "@openzeppelin/contracts/utils/ReentrancyGuard.sol";

contract OwnaFarmNFT is ERC1155, AccessControl, ReentrancyGuard {
    using SafeERC20 for IERC20;
    
    error InvoiceNotApproved();
    error InvoiceInactive();
    error ExceedsTarget();
    error AlreadyClaimed();
    error NotMature();
    error InvalidInvestment();
    error NotPending();
    
    bytes32 public constant ADMIN_ROLE = keccak256("ADMIN_ROLE");
    
    enum InvoiceStatus { Pending, Approved, Rejected, Funded, Completed }
    
    struct Invoice {
        address farmer;
        uint128 targetFund;
        uint128 fundedAmount;
        uint16 yieldBps;
        uint32 duration;
        uint32 createdAt;
        InvoiceStatus status;
        bytes32 offtakerId;
    }
    
    struct Investment {
        uint128 amount;
        uint32 tokenId;
        uint32 investedAt;
        bool claimed;
    }
    
    IERC20 public immutable GOLD;
    uint32 public nextTokenId = 1;
    
    mapping(uint256 => Invoice) public invoices;
    mapping(address => mapping(uint256 => Investment)) public investments;
    mapping(address => uint256) public investmentCount;
    
    uint256[] private _pendingIds;
    mapping(uint256 => uint256) private _pendingIndex;
    
    uint256[] private _availableIds;
    mapping(uint256 => uint256) private _availableIndex;
    
    event InvoiceSubmitted(uint256 indexed tokenId, address indexed farmer, bytes32 offtakerId, uint256 target);
    event InvoiceApproved(uint256 indexed tokenId, address indexed approver);
    event InvoiceRejected(uint256 indexed tokenId, address indexed rejector);
    event Invested(address indexed investor, uint256 indexed tokenId, uint256 amount, uint256 investmentId);
    event Harvested(address indexed investor, uint256 indexed investmentId, uint256 principal, uint256 yield);
    event InvoiceFullyFunded(uint256 indexed tokenId);
    
    constructor(address gold_) ERC1155("") {
        GOLD = IERC20(gold_);
        _grantRole(DEFAULT_ADMIN_ROLE, msg.sender);
        _grantRole(ADMIN_ROLE, msg.sender);
    }
    
    function submitInvoice(
        bytes32 offtakerId,
        uint128 targetFund,
        uint16 yieldBps,
        uint32 duration
    ) external returns (uint256 tokenId) {
        unchecked { tokenId = nextTokenId++; }
        
        invoices[tokenId] = Invoice({
            farmer: msg.sender,
            targetFund: targetFund,
            fundedAmount: 0,
            yieldBps: yieldBps,
            duration: duration,
            createdAt: uint32(block.timestamp),
            status: InvoiceStatus.Pending,
            offtakerId: offtakerId
        });
        
        _addToPending(tokenId);
        emit InvoiceSubmitted(tokenId, msg.sender, offtakerId, targetFund);
    }
    
    function approveInvoice(uint256 tokenId) external onlyRole(ADMIN_ROLE) {
        Invoice storage inv = invoices[tokenId];
        if (inv.status != InvoiceStatus.Pending) revert NotPending();
        
        inv.status = InvoiceStatus.Approved;
        _removeFromPending(tokenId);
        _addToAvailable(tokenId);
        
        emit InvoiceApproved(tokenId, msg.sender);
    }
    
    function rejectInvoice(uint256 tokenId) external onlyRole(ADMIN_ROLE) {
        Invoice storage inv = invoices[tokenId];
        if (inv.status != InvoiceStatus.Pending) revert NotPending();
        
        inv.status = InvoiceStatus.Rejected;
        _removeFromPending(tokenId);
        
        emit InvoiceRejected(tokenId, msg.sender);
    }
    
    function invest(uint256 tokenId, uint128 amount) external nonReentrant {
        Invoice storage inv = invoices[tokenId];
        if (inv.status != InvoiceStatus.Approved) revert InvoiceNotApproved();
        if (inv.fundedAmount + amount > inv.targetFund) revert ExceedsTarget();
        
        GOLD.safeTransferFrom(msg.sender, address(this), amount);
        unchecked { inv.fundedAmount += amount; }
        
        _mint(msg.sender, tokenId, 1, "");
        
        uint256 invId = investmentCount[msg.sender];
        investments[msg.sender][invId] = Investment({
            amount: amount,
            tokenId: uint32(tokenId),
            investedAt: uint32(block.timestamp),
            claimed: false
        });
        unchecked { investmentCount[msg.sender]++; }
        
        if (inv.fundedAmount >= inv.targetFund) {
            inv.status = InvoiceStatus.Funded;
            _removeFromAvailable(tokenId);
            emit InvoiceFullyFunded(tokenId);
        }
        emit Invested(msg.sender, tokenId, amount, invId);
    }
    
    function harvest(uint256 investmentId) external nonReentrant {
        Investment storage inv = investments[msg.sender][investmentId];
        if (inv.amount == 0) revert InvalidInvestment();
        if (inv.claimed) revert AlreadyClaimed();
        
        Invoice storage invoice = invoices[inv.tokenId];
        if (block.timestamp < inv.investedAt + invoice.duration) revert NotMature();
        
        inv.claimed = true;
        
        uint256 principal = inv.amount;
        uint256 yieldAmount;
        unchecked { yieldAmount = (principal * invoice.yieldBps) / 10000; }
        
        _burn(msg.sender, inv.tokenId, 1);
        GOLD.safeTransfer(msg.sender, principal + yieldAmount);
        
        emit Harvested(msg.sender, investmentId, principal, yieldAmount);
    }
    
    function _addToPending(uint256 tokenId) private {
        _pendingIds.push(tokenId);
        _pendingIndex[tokenId] = _pendingIds.length;
    }
    
    function _removeFromPending(uint256 tokenId) private {
        uint256 idx = _pendingIndex[tokenId];
        if (idx == 0) return;
        uint256 lastIdx = _pendingIds.length - 1;
        if (idx - 1 != lastIdx) {
            uint256 lastId = _pendingIds[lastIdx];
            _pendingIds[idx - 1] = lastId;
            _pendingIndex[lastId] = idx;
        }
        _pendingIds.pop();
        delete _pendingIndex[tokenId];
    }
    
    function _addToAvailable(uint256 tokenId) private {
        _availableIds.push(tokenId);
        _availableIndex[tokenId] = _availableIds.length;
    }
    
    function _removeFromAvailable(uint256 tokenId) private {
        uint256 idx = _availableIndex[tokenId];
        if (idx == 0) return;
        uint256 lastIdx = _availableIds.length - 1;
        if (idx - 1 != lastIdx) {
            uint256 lastId = _availableIds[lastIdx];
            _availableIds[idx - 1] = lastId;
            _availableIndex[lastId] = idx;
        }
        _availableIds.pop();
        delete _availableIndex[tokenId];
    }
    
    function getPendingInvoices(uint256 offset, uint256 limit) external view returns (uint256[] memory ids, Invoice[] memory data) {
        return _getInvoices(_pendingIds, offset, limit);
    }
    
    function getPendingCount() external view returns (uint256) {
        return _pendingIds.length;
    }
    
    function getAvailableInvoices(uint256 offset, uint256 limit) external view returns (uint256[] memory ids, Invoice[] memory data) {
        return _getInvoices(_availableIds, offset, limit);
    }
    
    function getAvailableCount() external view returns (uint256) {
        return _availableIds.length;
    }
    
    function _getInvoices(uint256[] storage arr, uint256 offset, uint256 limit) private view returns (uint256[] memory ids, Invoice[] memory data) {
        uint256 len = arr.length;
        if (offset >= len) return (new uint256[](0), new Invoice[](0));
        uint256 end = offset + limit;
        if (end > len) end = len;
        uint256 size = end - offset;
        ids = new uint256[](size);
        data = new Invoice[](size);
        for (uint256 i; i < size;) {
            uint256 id = arr[offset + i];
            ids[i] = id;
            data[i] = invoices[id];
            unchecked { ++i; }
        }
    }
    
    function getInvestment(address investor, uint256 investmentId) external view returns (Investment memory) {
        return investments[investor][investmentId];
    }
    
    function supportsInterface(bytes4 interfaceId) public view override(ERC1155, AccessControl) returns (bool) {
        return super.supportsInterface(interfaceId);
    }
    
    function setTokenURI(string calldata newUri) external onlyRole(ADMIN_ROLE) {
        _setURI(newUri);
    }
}